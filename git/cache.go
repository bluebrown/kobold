package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

type RepoCache struct {
	repos  map[string][]string
	mu     *sync.RWMutex
	dir    string
	cfetch *prometheus.CounterVec
}

func NewRepoCache(name string) *RepoCache {
	return &RepoCache{
		repos: make(map[string][]string),
		mu:    &sync.RWMutex{},
		dir:   filepath.Join(os.TempDir(), name+"-cache"),
	}
}

func (cache *RepoCache) SetCounter(cfetch *prometheus.CounterVec) {
	cache.cfetch = cfetch
}

func (cache *RepoCache) Fill(ctx context.Context, uris []PackageURI, lim int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	if err := os.MkdirAll(filepath.Join(cache.dir, "repos"), 0755); err != nil {
		return err
	}

	for _, uri := range uris {
		cache.repos[uri.Repo] = append(cache.repos[uri.Repo], uri.Ref)
	}

	for uri, refs := range cache.repos {
		cache.repos[uri] = unique(refs)
	}

	g := errgroup.Group{}
	g.SetLimit(lim)

	for uri, path := range cache.repos {
		uri, refs := uri, path
		g.Go(func() error {
			if cache.cfetch != nil {
				cache.cfetch.With(prometheus.Labels{"repo": uri}).Inc()
			}
			return Ensure(ctx, filepath.Join(cache.dir, "repos", uri), uri, refs...)
		})
	}

	err := g.Wait()

	return err
}

func (cache *RepoCache) Get(ctx context.Context, namespace, repo string) (string, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	src := filepath.Join(cache.dir, "repos", repo)

	d := filepath.Join(cache.dir, "namespaces", namespace, repo)

	if err := os.MkdirAll(d, 0755); err != nil {
		return "", fmt.Errorf("mkdir %q: %w", d, err)
	}

	// each caller gets a fresh copy of the repo. this will avoid collisions and
	// keep the source re-fetchable
	d = filepath.Join(d, uuid.NewString())

	if err := copy(ctx, src, d); err != nil {
		return "", fmt.Errorf("copy %q: %w", src, err)
	}
	return d, nil
}

func (cache *RepoCache) Purge(namespace string) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	return os.RemoveAll(filepath.Join(cache.dir, "namespaces", namespace))
}

func copy(ctx context.Context, src, dst string) error {
	cmd := exec.CommandContext(ctx, "cp", "-r", src, dst)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cp: %w: %s", err, string(b))
	}
	return nil
}

func unique(ss []string) []string {
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
