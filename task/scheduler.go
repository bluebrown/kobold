package task

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bluebrown/kobold/store/model"
)

// Scheduler wraps the pool and provides a scheduling interface events can be
// published to the Scheduler and will be scheduled to be processed by the pool.
// The Scheduler will debounce events to prevent overloading the pool. After the
// debounce interval has elapsed, the Scheduler will dispatch all pending tasks
// via the pool. New task resets the debounce interval.
type Scheduler struct {
	ingress chan *ingress
	pool    *Pool
}

func NewScheduler(ctx context.Context, q *model.Queries, size int, interval time.Duration) *Scheduler {
	return &Scheduler{
		// buffer incoming events to prevent blocking the caller,
		// incase the scheduler is currently blocking on pool.Dispatch()
		ingress: make(chan *ingress, 100),
		pool:    NewPool(ctx, size, q),
	}
}

func (s *Scheduler) SetHandler(h Handler) {
	s.pool.SetHandler(h)
}

// runs until error or the context passed to NewScheduler is canceled. will
// always wait for the pool to shutdown gracefully before returning
func (s *Scheduler) Run(debounce time.Duration) (err error) {
	// using named return values here in order to be able to join it in the
	// deferred cleanup function. The defer has access the value of err, set by
	// one of the switch cases. Additionally the deferred func may need to join
	// an error returned by the pool.

	defer func() {
		// we end up in this block by 2 conditions,
		// if either of schedule or dispatch returns an error
		// or if the context passed to NewScheduler is canceled

		// stop incomming messages
		close(s.ingress)

		// TODO: im am not sure if active tasks should be cancelled.
		// if there was an error, it is likely a database error, so
		// for now, we cancel out since database problems are dangerous
		s.pool.Cancel()

		err = errors.Join(err, s.pool.Wait())
		status := StatusSuccess
		if err != nil {
			status = StatusFailure
		}
		slog.InfoContext(s.pool.ctx, "scheduler shutdown completed", "status", status, "error", err)
	}()

	// t is used to debounce events. It is reset on every new event. once the
	// debounce interval has elapsed, the pending tasks are dispatched, and t is
	// set to nil. Since t.C is nil, the select case for t.C will not be
	// selected, and the loop will wait for new events. This prevents
	// over-polling the database for pending tasks. Additonally, events that
	// belong together are likely to be handled together, leading in fewer
	// commits to the git repository.
	t := new(time.Timer)

	slog.InfoContext(s.pool.ctx, "scheduler started", "debounce", debounce)

	for {
		select {
		case <-s.pool.Done():
			return
			// TODO: this doesnt need to be in the select. There is no race condition.
			// if s.Schedule call directly poool.Queue, we can also return the task id
			// as well as pontential decoding errors
		case ing := <-s.ingress:
			if err := s.schedule(ing); err != nil {
				if errors.Is(err, ErrNotDecodable) || errors.Is(err, ErrChannelNotFound) {
					continue
				}
				return fmt.Errorf("s.schedule(e): %w", err)
			}

			if t.C == nil {
				t = time.NewTimer(debounce)
				continue
			}

			if !t.Stop() {
				<-t.C
			}

			t.Reset(debounce)

		case <-t.C:
			if err := s.pool.Dispatch(); err != nil {
				return fmt.Errorf("s.pool.Dispatch(): %w", err)
			}
			t.C = nil
		}
	}
}

type ingress struct {
	Channel string
	Msg     []byte
}

func (s *Scheduler) Schedule(ctx context.Context, channel string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("trying to schedule event on canceled context: %w", err)
	}
	s.ingress <- &ingress{Channel: channel, Msg: data}
	return nil
}

func (s *Scheduler) schedule(e *ingress) error {
	return s.pool.Queue(s.pool.ctx, e.Channel, e.Msg)
}
