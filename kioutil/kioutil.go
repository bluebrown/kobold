package kioutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// a diff func takes two directory paths and returns a diff and a bool
// indicating if the diff is empty or not. It produces any in oder
// to interface with the message func
type DiffFunc func(string, string) (any, bool, error)

// a message func takes the diff and returns a commit message
// it takes an interface as input in order to allow the diff func
// to produce any type of diff
type MessageFunc func(data any) string

type GitPackageReadWriter struct {
	srcURI            GitPackageURI
	dstURI            GitPackageURI
	SetPathAnnotation bool
	diffFunc          DiffFunc
	diff              any
	msgFunc           MessageFunc
	ctx               context.Context
	msg               string
}

func NewGitPackageReadWriter(ctx context.Context, uri, dstBranch string) *GitPackageReadWriter {
	r := &GitPackageReadWriter{SetPathAnnotation: true}
	r.srcURI.MustUnmarshalText(uri)
	r.dstURI.MustUnmarshalText(uri)
	if dstBranch != "" {
		r.dstURI.Ref = dstBranch
	}
	return r
}

func (g GitPackageReadWriter) Read() ([]*yaml.RNode, error) {
	return (&GitPackageReader{
		srcURI:            g.srcURI,
		SetPathAnnotation: g.SetPathAnnotation,
		ctx:               g.ctx,
	}).Read()

}

func (g *GitPackageReadWriter) Write(nodes []*yaml.RNode) error {
	w := &GitPackageWriter{
		srcURI:   g.srcURI,
		dstURI:   g.dstURI,
		diffFunc: g.diffFunc,
		msgFunc:  g.msgFunc,
		ctx:      g.ctx,
	}
	if err := w.Write(nodes); err != nil {
		return err
	}
	g.diff = w.diff
	g.msg = w.msg
	return nil
}

func (g *GitPackageReadWriter) SetDiffFunc(fn DiffFunc) {
	g.diffFunc = fn
}

func (g GitPackageReadWriter) Diff() any {
	return g.diff
}

func (g GitPackageReadWriter) CommitMessage() string {
	return g.msg
}

func (g *GitPackageReadWriter) SetMsgFunc(fn MessageFunc) {
	g.msgFunc = fn
}

const (
	PathAnnotation = "kio.bluebrown.github.io/path"
)

type GitPackageReader struct {
	srcURI            GitPackageURI
	SetPathAnnotation bool
	ctx               context.Context
}

func NewGitPackageReader(srcURI string) *GitPackageReader {
	r := &GitPackageReader{}
	r.srcURI.MustUnmarshalText(srcURI)
	return r
}

func (g GitPackageReader) Read() ([]*yaml.RNode, error) {
	if g.ctx == nil {
		g.ctx = context.Background()
	}

	hostPath := mustTempDir()
	pkgPath := filepath.Join(hostPath, g.srcURI.Pkg)

	if err := gitE(g.ctx, "clone", "--branch", g.srcURI.Ref, "--depth", "1", g.srcURI.Repo, hostPath); err != nil {
		return nil, fmt.Errorf("clone repo: %w", err)
	}

	r := kio.LocalPackageReader{
		PackageFileName:    ".krmignore",
		IncludeSubpackages: true,
		PackagePath:        pkgPath,
	}

	if g.SetPathAnnotation {
		r.SetAnnotations = map[string]string{PathAnnotation: hostPath}
	} else {
		defer os.RemoveAll(hostPath)
	}

	nodes, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read package: %w", err)
	}

	return nodes, nil
}

func defaultMsg(data any) string {
	return fmt.Sprintf("chore(kioutil): update krm package\n\n%v\n", data)
}

type GitPackageWriter struct {
	srcURI   GitPackageURI
	dstURI   GitPackageURI
	diffFunc DiffFunc
	diff     any
	msgFunc  MessageFunc
	ctx      context.Context
	msg      string
}

func NewGitPackageWriter(dstURI, srcBranch string) *GitPackageWriter {
	r := &GitPackageWriter{}
	r.dstURI.MustUnmarshalText(dstURI)
	r.srcURI.MustUnmarshalText(dstURI)
	if srcBranch != "" {
		r.srcURI.Ref = srcBranch
	}
	return r
}

func (g *GitPackageWriter) Write(nodes []*yaml.RNode) error {
	if g.ctx == nil {
		g.ctx = context.Background()
	}

	var remoteHostPath string

	if len(nodes) > 0 {
		remoteHostPath = nodes[0].GetAnnotations()[PathAnnotation]
	}

	if remoteHostPath == "" {
		remoteHostPath = mustTempDir()
	}

	localHostPath := mustTempDir()
	localPkgPath := filepath.Join(localHostPath)

	defer func() {
		_ = os.RemoveAll(localHostPath)
		_ = os.RemoveAll(remoteHostPath)
	}()

	if err := os.MkdirAll(filepath.Dir(localPkgPath), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	err := (&kio.LocalPackageWriter{
		PackagePath:      localPkgPath,
		ClearAnnotations: []string{PathAnnotation},
	}).Write(nodes)

	if err != nil {
		return fmt.Errorf("write package: %w", err)
	}

	remotePkgPath := filepath.Join(remoteHostPath, g.dstURI.Pkg)

	if _, err := os.Stat(filepath.Join(remoteHostPath, ".git")); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(remoteHostPath), 0755); err != nil {
			return fmt.Errorf("create dir: %w", err)
		}
		if err := gitE(g.ctx, "clone", "--branch", g.srcURI.Ref, "--depth", "1", g.dstURI.Repo, remoteHostPath); err != nil {
			return fmt.Errorf("clone repo: %w", err)
		}
	}
	if g.diffFunc == nil {
		g.diffFunc = packageDiff
	}

	diff, changed, err := g.diffFunc(localPkgPath, remotePkgPath)
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}

	g.diff = diff

	if !changed {
		return nil
	}

	if g.dstURI.Ref != g.srcURI.Ref {
		if err := gitE(g.ctx, "-C", remoteHostPath, "checkout", "-b", g.dstURI.Ref); err != nil {
			return fmt.Errorf("checkout branch: %w", err)
		}
	}

	if err := copyutil.CopyDir(filesys.FileSystemOrOnDisk{}, localPkgPath, remotePkgPath); err != nil {
		return fmt.Errorf("sync file: %w", err)
	}

	if err := gitE(g.ctx, "-C", remoteHostPath, "add", "."); err != nil {
		return fmt.Errorf("add file: %w", err)
	}

	if g.msgFunc == nil {
		g.msgFunc = defaultMsg
	}

	g.msg = g.msgFunc(g.diff)

	if err := gitE(g.ctx, "-C", remoteHostPath, "commit", "-m", g.msg); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if err := gitE(g.ctx, "-C", remoteHostPath, "push", "--set-upstream", "origin", g.dstURI.Ref); err != nil {
		return fmt.Errorf("push: %w", err)
	}

	return nil
}

func git(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, string(b))
	}
	return string(b), nil
}

func gitE(ctx context.Context, args ...string) error {
	_, err := git(ctx, args...)
	return err
}

func packageDiff(a, b string) (any, bool, error) {
	var fs filesys.FileSystemOrOnDisk

	na := mustTempDir()
	defer os.RemoveAll(na)
	if err := copyutil.CopyDir(fs, a, na); err != nil {
		return "", false, fmt.Errorf("copy a: %w", err)
	}

	nb := mustTempDir()
	defer os.RemoveAll(nb)
	if err := copyutil.CopyDir(fs, b, nb); err != nil {
		return "", false, fmt.Errorf("copy b: %w", err)
	}

	// TODO: run na and nb through formatter to normalize whitespace

	cmd := exec.Command("git", "diff", "--no-index", "--minimal", nb, na)

	o, err := cmd.CombinedOutput()
	if err == nil {
		return string(o), false, nil
	}
	var ec *exec.ExitError
	if errors.As(err, &ec) {
		if ec.ExitCode() == 1 {
			return string(o), true, nil
		}
	}
	return "", false, fmt.Errorf("%s: %w: %s", "diff", err, string(o))
}

func VisitMapLeafs(nodes []*yaml.RNode, fn func(*yaml.MapNode) error) error {
	for _, node := range nodes {
		switch node.YNode().Kind {
		case yaml.SequenceNode:
			els, err := node.Elements()
			if err != nil {
				return err
			}
			if err := VisitMapLeafs(els, fn); err != nil {
				return err
			}
		case yaml.MappingNode:
			fields, err := node.Fields()
			if err != nil {
				return err
			}
			for _, field := range fields {
				f := node.Field(field)
				switch f.Value.YNode().Kind {
				case yaml.ScalarNode:
					if err := fn(f); err != nil {
						return err
					}
				case yaml.SequenceNode:
					els, err := f.Value.Elements()
					if err != nil {
						return err
					}
					if err := VisitMapLeafs(els, fn); err != nil {
						return err
					}
				default:
					if err := VisitMapLeafs([]*yaml.RNode{f.Value}, fn); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func CopyIO(r kio.Reader, w kio.Writer) error {
	nodes, err := r.Read()
	if err != nil {
		return fmt.Errorf("read nodes: %w", err)
	}
	err = w.Write(nodes)
	if err != nil {
		return fmt.Errorf("write nodes: %w", err)
	}
	return nil
}

func mustTempDir() string {
	dir, err := os.MkdirTemp("", "kioutil-*")
	if err != nil {
		panic(err)
	}
	return dir
}
