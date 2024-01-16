package config

import (
	"context"
	"fmt"

	"github.com/bluebrown/kobold/builtin"
	"github.com/bluebrown/kobold/store"
)

func ApplyBuiltins(ctx context.Context, q *store.Queries) error {
	for _, d := range builtin.Decoders() {
		if err := q.DecoderPut(ctx, store.DecoderPutParams{
			Name:   d.Name,
			Script: []byte(d.Script),
		}); err != nil {
			return fmt.Errorf("create decoder %q: %w", d.Name, err)
		}
	}

	for _, p := range builtin.PostHooks() {
		if err := q.PostHookPut(ctx, store.PostHookPutParams{
			Name:   p.Name,
			Script: []byte(p.Script),
		}); err != nil {
			return fmt.Errorf("create post hook %q: %w", p.Name, err)
		}
	}

	return nil
}
