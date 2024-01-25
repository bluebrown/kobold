package config

import (
	"context"
	"fmt"

	"github.com/bluebrown/kobold/plugin/builtin"
	"github.com/bluebrown/kobold/store/model"
)

func ApplyBuiltins(ctx context.Context, q *model.Queries) error {
	for _, d := range builtin.Decoders() {
		if err := q.DecoderPut(ctx, model.DecoderPutParams{
			Name:   d.Name,
			Script: []byte(d.Script),
		}); err != nil {
			return fmt.Errorf("create decoder %q: %w", d.Name, err)
		}
	}

	for _, p := range builtin.PostHooks() {
		if err := q.PostHookPut(ctx, model.PostHookPutParams{
			Name:   p.Name,
			Script: []byte(p.Script),
		}); err != nil {
			return fmt.Errorf("create post hook %q: %w", p.Name, err)
		}
	}

	return nil
}
