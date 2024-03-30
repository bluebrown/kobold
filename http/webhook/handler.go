package webhook

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

const warnQueryDeprecated = `sending channel name using query parameters is deprecated and will be removed in future releases. Please use path parameters instead.`

type scheduler interface {
	Schedule(ctx context.Context, chn string, data []byte) error
}

type Webhook struct {
	s scheduler
	r *mux.Router
}

func New(s scheduler) *Webhook {
	wh := &Webhook{s: s, r: mux.NewRouter()}
	wh.r.HandleFunc("/events/{chan}", wh.handleEvent)
	wh.r.HandleFunc("/events", wh.handleEvent).Queries("chan", "{chan}")
	return wh
}

func (api *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.r.ServeHTTP(w, r)
}

func (api *Webhook) handleEvent(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		http.Error(w, "unable to read body", http.StatusBadRequest)
		return
	}

	chn := mux.Vars(r)["chan"]
	logger := slog.With("chan", chn)

	if err := api.s.Schedule(r.Context(), chn, buf.Bytes()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		logger.Error("schedule task", "error", err)
		return
	}

	if !r.URL.Query().Has("chan") {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	logger.Warn("webhook event", "msg", warnQueryDeprecated)

	if _, err := w.Write([]byte(warnQueryDeprecated)); err != nil {
		logger.Error("write response body", "error", err)
	}
}
