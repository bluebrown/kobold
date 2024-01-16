package krm

import (
	"bytes"
	"context"
	"log/slog"
	"text/template"

	"github.com/bluebrown/kobold/kioutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

type Pipeline struct {
	RepoURI   string `json:"repoUri,omitempty"`
	SrcBranch string `json:"sourceBranch,omitempty"`
	DstBranch string `json:"destinationBranch,omitempty"`
}

func (opts Pipeline) Run(ctx context.Context, imageRefs []string) (msg string, changes, warnings []string, err error) {
	kf := NewImageRefUpdateFilter(nil, imageRefs...)

	grw := kioutil.NewGitPackageReadWriter(ctx, opts.RepoURI, opts.DstBranch)

	grw.SetDiffFunc(func(s1, s2 string) (any, bool, error) {
		return kf.Changes, len(kf.Changes) > 0, nil
	})

	grw.SetMsgFunc(func(data any) string {
		slog.InfoContext(ctx, "commit", "repo", opts.RepoURI, "branch", opts.DstBranch)
		msg := "chore(kobold): update krm package"
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, TemplateContext{
			Repo:     opts.RepoURI,
			Branch:   opts.DstBranch,
			Changes:  kf.Changes,
			Warnings: kf.Warnings,
		}); err != nil {
			slog.WarnContext(ctx, "failed to execute commit message template", "error", err)
			return msg
		}
		return buf.String()
	})

	pipe := kio.Pipeline{
		Inputs:  []kio.Reader{grw},
		Filters: []kio.Filter{kf},
		Outputs: []kio.Writer{grw},
	}

	if err := pipe.Execute(); err != nil {
		return "", nil, nil, err
	}

	return grw.CommitMessage(), kf.Changes, kf.Warnings, nil
}

type TemplateContext struct {
	Repo     string
	Branch   string
	Changes  []string
	Warnings []string
}

// TODO: this should be configurable
var tpl = template.Must(template.New("").Parse(`chore(kobold): update image refs
{{ range .Changes}}
- {{.}}
{{- end}}
{{ range .Warnings}}
- {{.}}
{{- end}}
`))
