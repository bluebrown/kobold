/* read a user friendly config file, and prepare the database accordingly. */
package config

import (
	"context"
	"fmt"

	"github.com/bluebrown/kobold/git"
	"github.com/bluebrown/kobold/store/model"
	"github.com/volatiletech/null/v8"
)

type Decoder struct {
	Name   string `toml:"name"`
	Script string `toml:"script"`
}

type Channel struct {
	Name    string `toml:"name"`
	Decoder string `toml:"decoder"`
}

type PostHook struct {
	Name   string `toml:"name"`
	Script string `toml:"script"`
}

type Pipeline struct {
	Name       string         `toml:"name"`
	RepoURI    git.PackageURI `toml:"repo_uri"`
	DestBranch string         `toml:"dest_branch"`
	Channels   []string       `toml:"channels"`
	PostHook   string         `toml:"post_hook"`
}

type Config struct {
	Version   string     `toml:"version"`
	Channels  []Channel  `toml:"channel"`
	Pipelines []Pipeline `toml:"pipeline"`
	PostHooks []PostHook `toml:"post_hook"`
	Decoders  []Decoder  `toml:"decoder"`
}

func (cfg *Config) Apply(ctx context.Context, q *model.Queries) error {
	for _, d := range cfg.Decoders {
		if err := q.DecoderPut(ctx, model.DecoderPutParams{
			Name:   d.Name,
			Script: []byte(d.Script),
		}); err != nil {
			return fmt.Errorf("create decoder %q: %w", d.Name, err)
		}
	}

	for _, p := range cfg.PostHooks {
		if err := q.PostHookPut(ctx, model.PostHookPutParams{
			Name:   p.Name,
			Script: []byte(p.Script),
		}); err != nil {
			return fmt.Errorf("create post hook %q: %w", p.Name, err)
		}
	}

	for _, c := range cfg.Channels {
		// ch := model.ChannelPutParams{Name: c.Name, DecoderName: store.NullString{String: c.Decoder, Valid: c.Decoder != ""}}
		ch := model.ChannelPutParams{Name: c.Name, DecoderName: null.NewString(c.Decoder, c.Decoder != "")}
		if err := q.ChannelPut(ctx, ch); err != nil {
			return fmt.Errorf("create channel %q: %w", c.Name, err)
		}
	}

	for _, p := range cfg.Pipelines {
		if err := q.PipelinePut(ctx, model.PipelinePutParams{
			Name:         p.Name,
			RepoUri:      p.RepoURI,
			DestBranch:   null.NewString(p.DestBranch, p.DestBranch != ""),
			PostHookName: null.NewString(p.PostHook, p.PostHook != ""),
		}); err != nil {
			return fmt.Errorf("create pipeline %q: %w", p.Name, err)
		}

		for _, c := range p.Channels {
			if err := q.SubscriptionPut(ctx, model.SubscriptionPutParams{
				PipelineName: p.Name,
				ChannelName:  c,
			}); err != nil {
				return fmt.Errorf("create subscription %q=>%q: %w", p.Name, c, err)
			}
		}
	}

	return nil
}
