package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

// the repo cache will keep a clean version of each repo fetched via Fill().
// Each time a repo is requested with Get(), a new copy of the local repo is
// made, and the path to it is returned. These copies are unique and persist in
// the given namespace until that namespace is purged. This allows multiple
// callers to use the same repo without worrying about collisions. It also
// allows the cache to refetch the repo, without having to reset branches or
// clean up after itself. The cached repositories contain only the .git
// directory, because they are obtained with git fetch. Therefore, after the
// caller has retrieved the path to the repo, they need to git switch to their
// desired branch. Repos and branches that have not been passed to Fill() will
// not be available. Calling Fill() with a repo that has already been cached
// will cause the cache to refetch the repo.
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

func (cache *RepoCache) Fill(ctx context.Context, lim int, uris ...PackageURI) error {
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
			if err := Ensure(ctx, filepath.Join(cache.dir, "repos", uri), uri, refs...); err != nil {
				return fmt.Errorf("ensure %q: %w", uri, err)
			}
			return nil
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
