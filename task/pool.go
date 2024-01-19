package task

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/volatiletech/null/v8"
	"golang.org/x/sync/errgroup"

	"github.com/bluebrown/kobold/dbutil"
	"github.com/bluebrown/kobold/store"
)

// pool implements a worker pool backed by the storage layer. Tasks are first
// queued, by calling the Queue() method, which stores them in the storage. Then
// task groups are dispatched to a pool of goroutines, by calling Dispatch().
// Task status is tracked, which allows to pause-resume the work or replay
// failed tasks. The pool will dispatch tasks to workers as they become
// available. Call Wait() to block until all tasks have been processed. If a
// task fails, it will be marked as failed in the database but the pool will
// continue to process other tasks. Only if there is an irecoverable error, the
// the pool drains remaining task and returns the error
type Pool struct {
	group      *errgroup.Group
	queries    *store.Queries
	ctx        context.Context
	handler    TaskHandler
	decoder    Decoder
	hookRunner HookRunner
	cancel     context.CancelFunc
	size       int
}

func NewPool(ctx context.Context, size int, queries *store.Queries) *Pool {
	ctx, cancel := context.WithCancel(ctx)
	eg, ectx := errgroup.WithContext(ctx)
	eg.SetLimit(size)
	return &Pool{
		ctx:        ectx,
		cancel:     cancel,
		group:      eg,
		queries:    queries,
		handler:    nil,
		decoder:    NewStarlarkDecoder(),
		hookRunner: NewStarlarkPostHook(),
		size:       size,
	}
}

func (p *Pool) SetHandler(h TaskHandler) {
	p.handler = h
}

// dispatch pending tasks. Will block until all task groups have been dispatched
func (p *Pool) Dispatch() error {
	if err := p.ctx.Err(); err != nil {
		return err
	}

	taskGroups, err := p.queries.TaskGroupsListPending(p.ctx)
	if err != nil {
		return err
	}

	// each dispatch gets its own cache, to avoid collisions
	cache := &repoCache{
		repos: make(map[string]string),
		base:  filepath.Join(os.TempDir(), "kobold-cache-"+uuid.NewString()),
	}

	// find all unique repos and clone them beforehand in order to avoid
	// exsessive cloning
	if err := cache.fill(p.ctx, taskGroups, p.size); err != nil {
		slog.WarnContext(p.ctx, "failed to fill cache", "error", err)
	}

	// the waitgroups is used to know when all task groups of this dispatch call
	// have been processed. This is used to purge the cache
	wg := sync.WaitGroup{}

	go func() {
		wg.Wait()
		cache.purge()
	}()

	for _, g := range taskGroups {
		g := g
		wg.Add(1)

		// NOTE, returning an error from this function provides a way to hault
		// the entire pool. Do not return an error here, unless you want to stop
		// the pool.
		p.group.Go(func() (err error) {
			defer wg.Done()

			ids := []string(g.TaskIds)

			swapped, err := p.queries.TaskGroupsStatusCompSwap(p.ctx, store.TaskGroupsStatusCompSwapParams{
				TaskGroupFingerprint: null.NewString(g.Fingerprint, true),
				Status:               string(StatusRunning),
				ReqStatus:            string(StatusPending),
				Ids:                  ids,
			})

			// if there is a database error, we consider the app irrecoverable
			if err != nil {
				return err
			}

			// swapped is the list of task ids that have been swapped from
			// pending. any id that is not in this list, shall not be handled by
			// this worker. for now, dont try to be smart and just bail out, if
			// the lists dont match
			if len(swapped) != len(ids) {
				return fmt.Errorf("attempt to set non %q task to %q: swapped=%v ids=%v",
					StatusPending, StatusRunning, swapped, ids)
			}

			slog.InfoContext(p.ctx, "task group dispatched", "fingerprint", g.Fingerprint)
			l := prometheus.Labels{"status": string(StatusRunning), "repo": g.RepoUri.Repo}
			metricRun.With(l).Inc()
			defer metricRun.With(l).Add(-1)

			var (
				status = StatusSuccess
				reason string
				warns  []string
			)

			if p.handler != nil {
				warns, err = p.handler(p.ctx, cache.get(g.RepoUri.Repo), g, p.hookRunner)
				// if the handler throws an error, we record the outcome but
				// continue to process other tasks
				if err != nil {
					status = StatusFailure
					reason = err.Error()
				}
			}

			swapped, err = p.queries.TaskGroupsStatusCompSwap(p.ctx, store.TaskGroupsStatusCompSwapParams{
				TaskGroupFingerprint: null.NewString(g.Fingerprint, true),
				Ids:                  ids,
				ReqStatus:            string(StatusRunning),
				Status:               string(status),
				FailureReason:        null.NewString(reason, reason != ""),
				Warnings:             dbutil.SliceText(warns),
			})

			slog.InfoContext(p.ctx, "task group done", "fingerprint", g.Fingerprint, "status", status)
			metricRun.With(prometheus.Labels{"status": string(status), "repo": g.RepoUri.Repo}).Add(1)

			if err != nil {
				return err
			}

			// the same as above. For the time being, we just bail out
			if len(swapped) != len(ids) {
				return fmt.Errorf("attempt to set non %q task to %q: swapped=%v ids=%v",
					StatusRunning, status, swapped, ids)
			}

			// since we used a named return value, to capture the error, make
			// sure we return nil here, so that no error is returned
			return nil
		})
	}

	return nil
}

func (p *Pool) Done() <-chan struct{} {
	return p.ctx.Done()
}

// waits for all current tasks to complete, then cancels the context. So the
// pool should be closed after calling this method
func (p *Pool) Wait() error {
	return p.group.Wait()
}

// cancels the context
func (p *Pool) Cancel() {
	p.cancel()
}

var (
	ErrChannelNotFound = fmt.Errorf("channel not found")
	ErrNotDecodable    = fmt.Errorf("not decodable")
)

func (p *Pool) Queue(ctx context.Context, channel string, msg []byte) (err error) {
	var dec []byte

	defer func() {
		slog.InfoContext(ctx, "task queued",
			"channel", channel,
			"defaultDecoder",
			len(dec) == 0, "error", err)

		metricMsgRecv.With(prometheus.Labels{
			"channel":  channel,
			"rejected": fmt.Sprintf("%v", err != nil)}).Inc()
	}()

	// fetch decoder info from db and decode data into a slice of imageRefs so
	// that the aggregated task handler does not need to know about the decoder
	dec, err = p.queries.ChannelDecoderGet(ctx, channel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %q", ErrChannelNotFound, channel)
		}
		return errors.Join(err, ErrNotDecodable)
	}

	var refs []string

	if dec == nil {
		refs = strings.Split(string(msg), "\n")
	} else {
		refs, err = p.decoder.Decode(channel, dec, msg)
		if err != nil {
			return errors.Join(ErrNotDecodable, err)
		}
	}

	_, err = p.queries.TasksAppend(ctx, store.TasksAppendParams{
		Msgs: dbutil.SliceText(refs),
		Name: channel,
	})

	return err
}

func (p *Pool) QueueReader(ctx context.Context, channel string, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if err := p.Queue(ctx, channel, scanner.Bytes()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

type repoCache struct {
	repos map[string]string
	base  string
	path  string
}

func (cache *repoCache) fill(ctx context.Context, gg []store.TaskGroup, lim int) error {

	// add a random dir inside the cache path to avoid collisions
	cache.path = cache.base

	if err := os.MkdirAll(cache.path, 0755); err != nil {
		return err
	}

	for _, g := range gg {
		cache.repos[g.RepoUri.Repo] = filepath.Join(cache.path, g.RepoUri.Repo)
	}

	g := errgroup.Group{}
	g.SetLimit(lim)

	for uri, path := range cache.repos {
		uri, path := uri, path
		g.Go(func() error {
			return cache.ensure(ctx, uri, path)
		})
	}

	err := g.Wait()

	return err
}

func (cache *repoCache) ensure(ctx context.Context, uri, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// since fetching leads to merge conflicts, we just delete the repo and
	// clone it again with depth 1
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove repo %q: %w", uri, err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", uri, path)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone %q: %w: %s", uri, err, string(b))
	}

	metricClone.With(prometheus.Labels{"repo": uri}).Inc()
	slog.InfoContext(ctx, "repo cloned", "repo", uri, "cache", path)

	return nil
}

func (cache *repoCache) get(repo string) string {
	path := cache.repos[repo]
	if path == "" {
		return ""
	}

	d := filepath.Join(cache.path, uuid.NewString())

	cmd := exec.Command("cp", "-r", path, d)
	if b, err := cmd.CombinedOutput(); err != nil {
		slog.Warn("failed to copy repo", "error", err, "output", string(b))
		return ""
	}

	return d
}

func (cache *repoCache) purge() error {
	return os.RemoveAll(cache.path)
}
