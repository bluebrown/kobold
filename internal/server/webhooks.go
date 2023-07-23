package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/bluebrown/kobold/internal/events"
	"github.com/bluebrown/kobold/internal/krm"
	"github.com/bluebrown/kobold/kobold/config"
)

func RequireHeaders(headers []config.Header, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range headers {
			if val := r.Header.Get(h.Key); val == "" || val != h.Value {
				log.Warn().Str("key", h.Key).Str("path", r.URL.Path).Msg("request does not contain required header")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		handler.ServeHTTP(w, r)
	})
}

func NewPushWebhook(id string, subs []chan events.PushData, ph events.PayloadHandler) http.Handler {
	logger := log.With().Str("endpoint", id).Logger()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info().Str("endpoint", id).Msg("push event received")

		var b bytes.Buffer
		if _, err := io.Copy(&b, r.Body); err != nil {
			logger.Debug().Err(err).Msg("could not copy body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		bodyBytes := b.Bytes()

		if err := ph.Validate(bodyBytes); err != nil {
			logger.Debug().Err(err).Msg("error while validation")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// since decoding may to external io and take some time,
		// do the work concurrently, since we know the payload is valid
		go func() {
			event, err := ph.Decode(bodyBytes)
			if err != nil {
				logger.Error().Err(err).Msg("failed to decode payload")
				return
			}
			logger.Info().
				Str("endpoint", id).
				Str("image", event.Image).
				Str("tag", event.Tag).
				Msg("dispatching event")
			for _, c := range subs {
				c <- event
			}
			logger.Debug().Msg("event dispatched to subscribers")
		}()

		// return 202 accepted, while the request is being processed
		// to avoid client timeout from the webhook sender
		w.WriteHeader(http.StatusAccepted)
	})
}

type commitMessenger interface {
	Msg(changes []krm.Change) (string, string, error)
}

type repoBot interface {
	Do(ctx context.Context, callback func(ctx context.Context, dir string) (title string, msg string, err error)) error
}

func NewSubscriber(id string, bot repoBot, renderer krm.Renderer, messenger commitMessenger) chan events.PushData {
	eventsChan := make(chan events.PushData, 10)
	logger := log.With().Str("subscriber", id).Logger()
	go func() {
		var (
			queue    = make([]events.PushData, 0, 100)
			debounce = new(time.Timer)
			delay    = time.Minute
		)
		for {
			select {
			case event := <-eventsChan:
				queue = append(queue, event)
				if debounce.C == nil {
					logger.Debug().Msg("queueing event")
					debounce = time.NewTimer(delay)
					continue
				}
				logger.Debug().Msg("debouncing event")
				if !debounce.Stop() {
					<-debounce.C
				}
				debounce.Reset(delay)
			case <-debounce.C:
				logger.Debug().Msg("processing queued events")
				if err := bot.Do(logger.WithContext(context.Background()), func(ctx context.Context, dir string) (string, string, error) {
					changes, err := renderer.Render(ctx, dir, queue)
					if err != nil {
						return "", "", fmt.Errorf("error while rendering: %w", err)
					}
					return messenger.Msg(changes)
				}); err != nil {
					logger.Error().Err(err).Msg("error while running bot")
				}
				queue = queue[:0]
				debounce.C = nil
			}
		}
	}()
	return eventsChan
}
