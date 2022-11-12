package krm

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/bluebrown/kobold/internal/events"
	"github.com/bluebrown/kobold/kobold"
)

type NopRenderer struct{}

func (NopRenderer) Render(ctx context.Context, dir string, events []events.PushData) ([]Change, error) {
	return nil, nil
}

// the image node handler func is responsible for handling the actual image nodes found
// be the resolver. It may mutate the image ref or do nothing
type ImageNodeHandlerFunc func(source, parent string, imgNode *yaml.MapNode) error

// the resolver is resonsible for finding one or more image node in a given yaml document
type Resolver func(node *yaml.RNode, source string, handleImage ImageNodeHandlerFunc) error

// the resolver selector should return the correct resolver based on the file
// for example for a docker-compose.yaml, the compose resolver should be returned
type ResolverSelector func(ctx context.Context, source string) Resolver

// the renderer is the high level struct used with the krm framework.
// its render function runs a kio pipeline using a custom filter based
// on the renderer options
type renderer struct {
	skipfn           kio.LocalPackageSkipFileFunc
	selector         ResolverSelector
	defaultRegistry  string
	imageNodeHandler *ImageNodeHandler
	writer           kio.Writer
}

type RendererOption func(r *renderer)

// scope this renderer to the list of glob pattern
func WithScopes(scopes []string) RendererOption {
	return func(r *renderer) {
		if len(scopes) > 0 {
			r.skipfn = ignore(scopes)
		}
	}
}

// the selector determines which resolver to use for a given file name
func WithSelector(selector ResolverSelector) RendererOption {
	return func(r *renderer) {
		r.selector = selector
	}
}

// the default registry will be used for any image that has no
// fully qualified domain name
func WithDefaultRegistry(registry string) RendererOption {
	return func(r *renderer) {
		r.defaultRegistry = registry
	}
}

// the imageref template is used to format the new image ref
// when updating image nodes.
func WithImagerefTemplate(t string) RendererOption {
	return func(r *renderer) {
		r.imageNodeHandler = NewImageNodeHandler(t)
	}
}

func WithWriter(w kio.Writer) RendererOption {
	return func(r *renderer) {
		r.writer = w
	}
}

// create a new renderer with the given options
func NewRenderer(opts ...RendererOption) renderer {
	r := renderer{}

	for _, o := range opts {
		o(&r)
	}

	if r.selector == nil {
		r.selector = NewSelector(kobold.DefaultAssociations)
	}

	if r.defaultRegistry == "" {
		r.defaultRegistry = name.DefaultRegistry
	}

	if r.imageNodeHandler == nil {
		r.imageNodeHandler = NewImageNodeHandler(DefaultImagerefTemplate)
	}

	r.imageNodeHandler.AddNameOptions(name.WithDefaultRegistry(r.defaultRegistry))

	return r
}

// Render takes an input directory path and a slice of events as arguments.
// It uses a kio pipeline to walk the directory and potentially mutates image references
// based on the given events. It will report any changes it made back
func (r renderer) Render(ctx context.Context, dir string, events []events.PushData) ([]Change, error) {
	l := zerolog.Ctx(ctx)
	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("dir", dir)
	})

	log.Ctx(ctx).Trace().Msg("rendering")

	rw := &kio.LocalPackageReadWriter{
		PackagePath:        dir,
		PreserveSeqIndent:  true,
		PackageFileName:    PackageFile,
		IncludeSubpackages: true,
		WrapBareSeqNode:    true,
		NoDeleteFiles:      true,
		FileSkipFunc:       r.skipfn,
	}

	f := filter{
		context:  ctx,
		Events:   events,
		selector: r.selector,
		handler:  r.imageNodeHandler,
	}

	if r.writer == nil {
		r.writer = rw
	}

	err := kio.Pipeline{
		Inputs:  []kio.Reader{rw},
		Filters: []kio.Filter{&f},
		Outputs: []kio.Writer{r.writer},
	}.Execute()

	return f.Changes, err
}

// the filter is used to implement the kio.Filter interface
// it is initialized and invoked by the higher level renderer
type filter struct {
	context  context.Context
	Events   []events.PushData
	Changes  []Change
	selector ResolverSelector
	handler  *ImageNodeHandler
}

func (fn *filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	// the changes are used to capture any change that has been made
	// while running the filter.
	// They are returned by the Renderer wrapping the filter
	fn.Changes = make([]Change, 0)

	// each node represents a file
	for _, node := range nodes {
		// the filename is read from the path annotation, added by the krm framework
		source := node.GetAnnotations()[kioutil.PathAnnotation]
		log.Ctx(fn.context).Trace().Str("source", source).Msg("handling node")

		// select the resolver based on the source and resolve the image nodes with it
		// once an image node is found, the imageNodeHandler is invoked
		resolver := fn.selector(fn.context, source)
		if resolver == nil {
			log.Warn().Str("source", source).Msg("no matching selector")
			continue
		}
		if err := resolver(node, source, fn.imageNodeHandler); err != nil {
			log.Ctx(fn.context).Warn().Err(err).Str("source", source).Msg("failed to handle node")
		}
	}

	// return all nodes once done, in order to not delete the files
	return nodes, nil
}

func (fn *filter) imageNodeHandler(source, parent string, imgNode *yaml.MapNode) error {
	changed, change, err := fn.handler.HandleImageNode(imgNode, fn.Events)
	if err != nil {
		return fmt.Errorf("failed to handle image node: %w", err)
	}
	if !changed {
		log.Ctx(fn.context).Trace().
			Str("source", source).
			Str("parent", parent).
			Msg("no change")
		return nil
	}
	change.Source = source
	change.Parent = parent
	fn.Changes = append(fn.Changes, change)
	log.Ctx(fn.context).Trace().
		Str("source", source).
		Str("parent", parent).
		Str("old", change.OldImageRef).
		Str("new", change.NewImageRef).
		Msg("change")
	return nil
}

const (
	KubeFieldContainerImage = "image"
	KubeFieldContainerName  = "name"
)

func resolveKube(node *yaml.RNode, source string, handleImage ImageNodeHandlerFunc) error {
	containers, err := node.Pipe(yaml.LookupFirstMatch(yaml.ConventionalContainerPaths))
	if err != nil {
		return err
	}
	if containers == nil {
		return err
	}
	return containers.VisitElements(func(container *yaml.RNode) error {
		imgNode := container.Field(KubeFieldContainerImage)
		cname, err := container.GetString(KubeFieldContainerName)
		if err != nil {
			cname = "unset"
		}
		return handleImage(source, cname, imgNode)
	})
}

const (
	ComposeFieldServices = "services"
	ComposeFieldImage    = "image"
)

func resolveCompose(node *yaml.RNode, source string, handleImage ImageNodeHandlerFunc) error {
	svcNode := node.Field(ComposeFieldServices)
	if svcNode == nil {
		return nil
	}
	return svcNode.Value.VisitFields(func(n *yaml.MapNode) error {
		imgNode := n.Value.Field(ComposeFieldImage)
		if imgNode == nil {
			return nil
		}
		return handleImage(source, yaml.GetValue(n.Key), imgNode)
	})
}

const (
	KoFieldDefaultBaseImage   = "defaultBaseImage"
	KoFieldBaseImageOverrides = "baseImageOverrides"
)

func resolveKo(node *yaml.RNode, source string, handleImage ImageNodeHandlerFunc) error {
	if imgNode := node.Field(KoFieldDefaultBaseImage); imgNode != nil {
		if err := handleImage(source, KoFieldDefaultBaseImage, imgNode); err != nil {
			return err
		}
	}
	imgMap := node.Field(KoFieldBaseImageOverrides)
	if imgMap == nil {
		return nil
	}
	return imgMap.Value.VisitFields(func(n *yaml.MapNode) error {
		return handleImage(source, yaml.GetValue(n.Key), n)
	})
}

func NewSelector(fa []kobold.FileTypeSpec) ResolverSelector {
	return func(ctx context.Context, source string) Resolver {
		base := filepath.Base(source)
		var res Resolver
		for _, a := range fa {
			ok, err := filepath.Match(a.Pattern, base)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("failed to match filetype")
				continue
			}
			if ok {
				res = lookupResolver(a.Kind)
				break
			}
		}
		return res
	}
}

func lookupResolver(kind kobold.FileTypeKind) Resolver {
	switch kind {
	case kobold.FileTypeKubernetes:
		return resolveKube
	case kobold.FileTypeCompose:
		return resolveCompose
	case kobold.FileTypeKo:
		return resolveKo
	}
	return nil
}
