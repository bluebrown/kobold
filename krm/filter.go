package krm

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func DefaultNodeHandler(_, currentRef, nextRef string, opts Options) (string, error) {
	if currentRef == nextRef {
		return currentRef, nil
	}

	fullRef := currentRef
	if opts.Part == "tag" {
		fullRef = fmt.Sprintf("%s:%s", opts.Context, currentRef)
	}
	oldRef, err := name.ParseReference(fullRef)
	if err != nil {
		return currentRef, err
	}

	newRef, _, err := ParseImageRefWithDigest(nextRef)
	if err != nil {
		return currentRef, err
	}

	if oldRef.Context().Name() != newRef.Context().Name() {
		return currentRef, nil
	}

	ok, err := MatchTag(newRef.Identifier(), opts)
	if err != nil {
		return currentRef, err
	}

	if !ok {
		return currentRef, nil
	}

	if _, err := name.ParseReference(nextRef); err != nil {
		return currentRef, err
	}

	if opts.Part == "tag" {
		return newRef.Identifier(), nil
	}
	return nextRef, nil
}

var CommentPrefix = "# kobold:"

type NodeHandler func(key, currentRef, nextRef string, opts Options) (string, error)

type ImageRefUpdateFilter struct {
	handler   NodeHandler
	imageRefs []string
	Changes   []Change
	Warnings  []string
}

type Change struct {
	Description string
	Repo        string
}

// create a new krm filter. The filter will traverse all nodes and invoke the
// handler if a map node with a line comment matching the CommentPrefix is
// found. The handler is responsible for determining if the node should be
// updated or not and what the new value should be. If no handler is passed, a
// default handler will be used. The image refs passed are new images references
// that may replace the current image ref. For any found map node, the handler
// will be invoked, once for each passed image ref.
func NewImageRefUpdateFilter(handler NodeHandler, imageRefs ...string) *ImageRefUpdateFilter {
	if handler == nil {
		handler = DefaultNodeHandler
	}
	return &ImageRefUpdateFilter{handler: handler, imageRefs: imageRefs}
}

func (i *ImageRefUpdateFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	// TODO: use top level loop to capture the current source file info.
	err := VisitMapLeafs(nodes, func(mn *yaml.MapNode) error {
		lineComment := mn.Value.YNode().LineComment

		if !strings.HasPrefix(lineComment, CommentPrefix) {
			return nil
		}

		opts, err := ParseOpts(strings.TrimPrefix(lineComment, CommentPrefix))
		if err != nil {
			i.Warnings = append(i.Warnings, fmt.Sprintf("failed to parse options: %v", err))
			return nil
		}

		originalValue := mn.Value.YNode().Value

		for _, imageRef := range i.imageRefs {
			v, err := i.handler(mn.Key.YNode().Value, mn.Value.YNode().Value, imageRef, opts)
			if err != nil {
				i.Warnings = append(i.Warnings, fmt.Sprintf("failed to update image ref %q: %v", imageRef, err))
				continue
			}
			mn.Value.YNode().Value = v
		}
		newValue := mn.Value.YNode().Value
		if originalValue != newValue {
			description := fmt.Sprintf("%s: %q -> %q", mn.Key.YNode().Value, originalValue, newValue)
			repo := GetRepoName(newValue)
			change := Change{description, repo}
			i.Changes = append(i.Changes, change)
		}
		return nil
	})
	return nodes, err
}

func GetRepoName(image string) (result string) {
	result = ""
	s := strings.LastIndex(image, "/")
	newS := image[s+1:]
	e := strings.Index(newS, ":")
	if e == -1 {
		return newS
	}
	result = newS[:e]
	return result
}

func ParseImageRefWithDigest(s string) (name.Reference, string, error) {
	rawRef, digest, _ := strings.Cut(s, "@")
	// NOTE: not checking ok here, to allow the user to use refs without digest
	// it is up to them to decide if that is acceptable or not.

	tag, err := name.NewTag(rawRef, name.StrictValidation)
	if err != nil {
		return nil, "", fmt.Errorf("invalid tag: %w", err)
	}

	return tag, digest, nil
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
