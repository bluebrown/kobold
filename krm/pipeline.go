package krm

import (
	"context"
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/kio"
)

func Pipeline(ctx context.Context, pkg string, refs ...string) ([]Change, []string, error) {
	rw := &kio.LocalPackageReadWriter{
		PackageFileName:     ".krmignore",
		PackagePath:         pkg,
		WrapBareSeqNode:     true,
		IncludeSubpackages:  true,
		PreserveSeqIndent:   true,
		NoDeleteFiles:       true,
		ErrorIfNonResources: false,
	}

	filter := NewImageRefUpdateFilter(nil, refs...)

	pipe := kio.Pipeline{
		Inputs:  []kio.Reader{rw},
		Filters: []kio.Filter{filter},
		Outputs: []kio.Writer{rw},
	}

	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	if err := pipe.Execute(); err != nil {
		return nil, nil, fmt.Errorf("kio pipeline: %w", err)
	}

	return filter.Changes, filter.Warnings, nil
}
