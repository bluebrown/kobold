package webhook

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/bluebrown/kobold/task"
	"github.com/gorilla/mux"
)

type Webhook struct {
	s *task.Scheduler
}

func New(s *task.Scheduler) *Webhook {
	return &Webhook{
		s: s,
	}
}

func (api *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		http.Error(w, "unable to read body", http.StatusBadRequest)
		return
	}

	muxVars := mux.Vars(r)
	channelName := muxVars["chan"]

	if err := api.s.Schedule(r.Context(), channelName, buf.Bytes()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Log and send back deprecation notice in response, if channel name is being sent using query
	// parameter.
	if r.URL.Query().Has("chan") {
		slog.Warn("Sending channel name using query parameters is deprecated")

		_, err := w.Write([]byte("Deprecated API: Send channel name using path parameter instead of query parameter"))
		if err != nil {
			slog.Error("write response body", "error", err)
		}
	}

	w.WriteHeader(http.StatusAccepted)
}
