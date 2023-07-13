package krm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/semver"
	"github.com/bluebrown/kobold/internal/events"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/google/go-containerregistry/pkg/name"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Renderer interface {
	Render(ctx context.Context, dir string, events []events.PushData) ([]Change, error)
}

const (
	PackageFile             = "kobold.yaml"
	CommentPrefix           = "# kobold: "
	DefaultImagerefTemplate = "{{ .Image }}:{{ .Tag }}@{{ .Digest }}"
)

const (
	TypeExact  = "exact"
	TypeSemver = "semver"
	TypeRegex  = "regex"
)

const (
	KeyType = "type"
	KeyTag  = "tag"
)

type Change struct {
	Source            string
	Parent            string
	OldImageRef       string
	NewImageRef       string
	OptionsExpression string
}

type Options struct {
	Type string
	Tag  string
}

func ParseOpts(s string) (Options, error) {
	kvs := strings.Split(s, ";")
	opts := Options{}
	for _, kv := range kvs {
		parts := strings.SplitN(kv, ":", 2)
		if len(parts) != 2 {
			return Options{}, fmt.Errorf("invalid key value pair: %s", kv)
		}
		switch strings.TrimSpace(parts[0]) {
		case KeyType:
			opts.Type = strings.TrimSpace(parts[1])
		case KeyTag:
			opts.Tag = strings.TrimSpace(parts[1])
		default:
			return Options{}, fmt.Errorf("invalid key: %s", parts[0])
		}
	}
	return opts, nil
}

// check if the provided tag matches per options.
// Exact, semver or regex
func MatchTag(tag string, opts Options) (bool, error) {
	switch opts.Type {
	case TypeExact:
		if tag != opts.Tag {
			return false, nil
		}
	case TypeSemver:
		c, err := semver.NewConstraint(opts.Tag)
		if err != nil {
			return false, fmt.Errorf("could not parse version constraint from opt: %w", err)
		}
		v, err := semver.NewVersion(tag)
		if err != nil || !c.Check(v) {
			return false, nil
		}
	case TypeRegex:
		ok, err := regexp.Match(fmt.Sprintf("^%s$", opts.Tag), []byte(tag))
		if err != nil {
			return false, fmt.Errorf("invalid regex in opt: %w", err)
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// The ImageNodeHandler is used to encasulate the logic
// of handling update nodes. It holds a template to avoid
// creating it on each run new
type ImageNodeHandler struct {
	tpl  *template.Template
	opts []name.Option
}

// FIXME: this panics on invalid template. There should be a better way
// since this function will not only run at program start
func NewImageNodeHandler(t string, nameOptions ...name.Option) *ImageNodeHandler {
	return &ImageNodeHandler{
		tpl:  template.Must(template.New("").Parse(t)),
		opts: nameOptions,
	}
}

func (h *ImageNodeHandler) AddNameOptions(opts ...name.Option) {
	h.opts = append(h.opts, opts...)
}

// Check if the given map node is eligible for an image update and
// update the image of so.
// The check is based on an inline comment in
// the form of:
//
// kobold: tag: [tag|semver-constraint|regex-pattern]; type: [exact|semver|regex]
//
// If the image has been updated, change data is returned.
func (h *ImageNodeHandler) HandleImageNode(imgNode *yaml.MapNode, events []events.PushData) (bool, Change, error) {
	lineComment := imgNode.Value.YNode().LineComment
	if lineComment == "" || !strings.HasPrefix(lineComment, CommentPrefix) {
		return false, Change{}, nil
	}

	oldRef, err := name.ParseReference(imgNode.Value.YNode().Value, h.opts...)
	if err != nil {
		return false, Change{}, err
	}

	optsExpr := strings.TrimPrefix(lineComment, CommentPrefix)
	opts, err := ParseOpts(optsExpr)
	if err != nil {
		return false, Change{}, err
	}

	var change Change

	// since the events are expected to be ordered by time
	// later events will overwrite previous events if they
	// match the same image, hence only a single change is emitted
	// for a given image node
	for _, event := range events {
		if oldRef.Context().String() != event.Image {
			continue
		}

		ok, err := MatchTag(event.Tag, opts)
		if err != nil {
			return false, Change{}, fmt.Errorf("match error: %w", err)
		}
		if !ok {
			continue
		}

		// renderer the new image ref using the template
		w := new(bytes.Buffer)
		if err := h.tpl.Execute(w, event); err != nil {
			return false, Change{}, fmt.Errorf("failed to render image ref: %w", err)
		}

		// parse the new ref to ensure it is valid
		newRef, err := name.ParseReference(w.String(), h.opts...)
		if err != nil {
			return false, Change{}, fmt.Errorf("new image ref is not valid: %w", err)
		}

		// if the new ref is the same as the old one
		// don't report a change.
		// this happens if the tag has matched the
		// event but the ref was already correct
		// NOTE, the not normalized forms are used
		// for this comparison as to not triger
		// false positives. For example parse ref
		// will remove reduntant tags
		if oldRef.String() == newRef.String() {
			continue
		}

		imgNode.Value.YNode().SetString(newRef.String())

		change = Change{
			OldImageRef:       oldRef.String(),
			NewImageRef:       newRef.String(),
			OptionsExpression: optsExpr,
		}
	}

	// if the change struct is not empty, we report
	// that we changed something
	return change.NewImageRef != "", change, nil
}

// returns a function that can be used as skipFn for a kio.Reader
func ignore(scopes []string) func(relPath string) bool {
	if len(scopes) < 1 {
		return nil
	}
	patterns := make([]gitignore.Pattern, len(scopes))
	for i, path := range scopes {
		patterns[i] = gitignore.ParsePattern(path, nil)
	}
	matcher := gitignore.NewMatcher(patterns)
	sep := string(os.PathSeparator)
	return func(relPath string) bool {
		// TODO: why is isDir always false here?
		return !matcher.Match(strings.Split(relPath, sep), false)
	}
}
