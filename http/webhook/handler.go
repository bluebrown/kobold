package webhook

import (
	"bytes"
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

	var channelName string

	// Try to get the channel name from the path parameter. If not found in the path parameter, then
	// get it from the query parameter.
	muxVars := mux.Vars(r)
	channelName = muxVars["chan"]
	if len(channelName) == 0 {
		channelName = r.URL.Query().Get("chan")
	}

	if err := api.s.Schedule(r.Context(), channelName, buf.Bytes()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
