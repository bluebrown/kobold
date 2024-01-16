package webhook

import (
	"bytes"
	"net/http"

	"github.com/bluebrown/kobold/task"
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
	if err := api.s.Schedule(r.Context(), r.URL.Query().Get("chan"), buf.Bytes()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
