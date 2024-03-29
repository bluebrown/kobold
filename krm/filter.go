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

	oldRef, err := name.ParseReference(currentRef)
	if err != nil {
		return currentRef, err
	}

	newRef, _, err := parseImageRefWithDigest(nextRef)
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

	return nextRef, nil
}

var CommentPrefix = "# kobold:"

type NodeHandler func(key, currentRef, nextRef string, opts Options) (string, error)

type ImageRefUpdateFilter struct {
	handler   NodeHandler
	imageRefs []string
	Changes   []string
	Warnings  []string
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

		if originalValue != mn.Value.YNode().Value {
			i.Changes = append(i.Changes, fmt.Sprintf("%s: %q -> %q", mn.Key.YNode().Value, originalValue, mn.Value.YNode().Value))
		}

		return nil
	})

	return nodes, err
}

func parseImageRefWithDigest(s string) (name.Reference, string, error) {
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
