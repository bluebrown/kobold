package task

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bluebrown/kobold/store"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

type repoCache struct {
	repos map[string][]string
	cache string
	tmp   string
}

func (cache *repoCache) fill(ctx context.Context, gg []store.TaskGroup, lim int) error {

	if err := os.MkdirAll(cache.cache, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(cache.tmp, 0755); err != nil {
		return err
	}

	for _, g := range gg {
		cache.repos[g.RepoUri.Repo] = append(cache.repos[g.RepoUri.Repo], g.RepoUri.Ref)
	}

	g := errgroup.Group{}
	g.SetLimit(lim)

	for uri, path := range cache.repos {
		uri, refs := uri, path
		g.Go(func() error {
			return cache.ensure(ctx, uri, dedupe(refs))
		})
	}

	err := g.Wait()

	return err
}

func (cache *repoCache) ensure(ctx context.Context, uri string, refs []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	path := filepath.Join(cache.cache, uri)

	if _, err := os.Stat(path); err != nil {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("remove %q: %w", path, err)
		}

		slog.InfoContext(ctx, "cloning repo", "repo", uri)

		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("mkdir %q: %w", path, err)
		}

		if err := run(ctx, "git", "init", path); err != nil {
			return fmt.Errorf("git init %q: %w", uri, err)
		}

		if err := run(ctx, "git", "-C", path, "remote", "add", "origin", uri); err != nil {
			return fmt.Errorf("git remote add %q: %w", uri, err)
		}
	} else {
		slog.InfoContext(ctx, "updating repo", "repo", uri)
	}

	args := []string{"git", "-C", path, "remote", "set-branches", "origin"}
	args = append(args, refs...)
	if err := run(ctx, args...); err != nil {
		return fmt.Errorf("git remote set-branches %q: %w", uri, err)
	}

	if err := run(ctx, "git", "-C", path, "fetch", "--depth", "1"); err != nil {
		return fmt.Errorf("git fetch %q: %w", uri, err)
	}

	metricGitFetch.With(prometheus.Labels{"repo": uri}).Inc()

	return nil
}

func (cache *repoCache) get(repo string) string {
	src := filepath.Join(cache.cache, repo)
	d := filepath.Join(cache.tmp, uuid.NewString())
	if err := run(context.Background(), "cp", "-r", src, d); err != nil {
		slog.Warn("failed to copy repo", "repo", repo, "error", err)
		return ""
	}
	return d
}
func (cache *repoCache) cleanTmp() error {
	return os.RemoveAll(cache.tmp)
}

func run(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, string(b))
	}
	return nil
}

func dedupe(ss []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, s := range ss {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
