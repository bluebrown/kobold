package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/bluebrown/kobold/api/docs"
	"github.com/bluebrown/kobold/store"
)

// @license.name	BSD-3-Clause
type WebAPI struct {
	q      *store.Queries
	router *mux.Router
}

// create a new web api handler. Requires to know the basepath its being served
// on, in order to generate correct swagger docs. It will not register routes on
// the basepath, the caller should remove the basepath from the mux before
// calling ServeHTTP
func New(basepath string, q *store.Queries) *WebAPI {
	api := WebAPI{q, mux.NewRouter()}

	docs.SwaggerInfo.Title = "Kobold API"
	docs.SwaggerInfo.Version = "dev"
	docs.SwaggerInfo.BasePath = basepath

	api.router.PathPrefix("/docs/").Handler(http.StripPrefix("/docs", httpSwagger.Handler())).Methods("GET")
	api.router.Path("/docs").Handler(http.RedirectHandler(basepath+"/docs/", http.StatusMovedPermanently)).Methods("GET")

	api.router.HandleFunc("/channels", api.GetChannelList).Methods("GET")
	api.router.HandleFunc("/channels/{name}", api.GetChannel).Methods("GET")

	api.router.HandleFunc("/decoders", api.GetDecoderList).Methods("GET")
	api.router.HandleFunc("/decoders/{name}", api.GetDecoder).Methods("GET")

	api.router.HandleFunc("/pipelines", api.GetPipelineList).Methods("GET")
	api.router.HandleFunc("/pipelines/{name}", api.GetPipeline).Methods("GET")
	api.router.HandleFunc("/pipelines/{name}/runs", api.GetPipelineRunList).Methods("GET")

	api.router.HandleFunc("/posthooks", api.GetPostHookList).Methods("GET")
	api.router.HandleFunc("/posthooks/{name}", api.GetPostHook).Methods("GET")

	api.router.HandleFunc("/tasks", api.GetTaskList).Methods("GET")
	api.router.HandleFunc("/tasks/{name}", api.GetTask).Methods("GET")

	api.router.HandleFunc("/runs", api.GetRunList).Methods("GET")
	api.router.HandleFunc("/runs/{name}", api.GetRun).Methods("GET")

	return &api
}

var decoder = schema.NewDecoder()

var stati = []string{"pending", "running", "success", "failure"}

func (api *WebAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}

func (api *WebAPI) send(w http.ResponseWriter, r *http.Request, code int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

type errorMsg struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (api *WebAPI) error(w http.ResponseWriter, r *http.Request, code int) {
	w.WriteHeader(code)
	api.send(w, r, code, errorMsg{code, http.StatusText(code)})
}

func (api *WebAPI) respond(w http.ResponseWriter, r *http.Request, data any, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		api.error(w, r, http.StatusNotFound)
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		api.error(w, r, http.StatusInternalServerError)
		return
	}
	if data == nil {
		api.error(w, r, http.StatusNotFound)
		return
	}
	api.send(w, r, http.StatusOK, data)
}

// GetChannel godoc
//
//	@Router		/channels/{name} [get]
//	@Summary	get a channel by name
//	@Tags		channels
//	@Produce	json
//	@Param		name	path		string	true	"channel name"
//	@Success	200		{object}	store.Channel
//	@Response	default	{object}	errorMsg "Error"
func (api *WebAPI) GetChannel(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	d, err := api.q.ChannelGet(r.Context(), name)
	api.respond(w, r, d, err)
}

// GetChannelList godoc
//
//	@Router		/channels [get]
//	@Summary	get a list of channels
//	@Tags		channels
//	@Produce	json
//	@Success	200		{array}		store.Channel
//	@Response	default	{object}	errorMsg "Error"
func (api *WebAPI) GetChannelList(w http.ResponseWriter, r *http.Request) {
	d, err := api.q.ChannelList(r.Context())
	api.respond(w, r, d, err)
}

// GetDecoder godoc
//
//	@Router		/decoders/{name} [get]
//	@Summary	get a decoder by name
//	@Tags		decoders
//	@Produce	json
//	@Param		name	path		string	true	"decoder name"
//	@Success	200		{object}	store.Decoder
//	@Response	default	{object}	errorMsg "Error"
func (api *WebAPI) GetDecoder(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	d, err := api.q.DecoderGet(r.Context(), name)
	api.respond(w, r, d, err)
}

// GetDecoderList godoc
//
//	@Router		/decoders [get]
//	@Summary	get a list of decoders
//	@Tags		decoders
//	@Produce	json
//	@Success	200		{array}		store.Decoder
//	@Response	default	{object}	errorMsg "Error"
func (api *WebAPI) GetDecoderList(w http.ResponseWriter, r *http.Request) {
	d, err := api.q.DecoderList(r.Context())
	api.respond(w, r, d, err)
}

// GetPipeline godoc
//
//	@Router		/pipelines/{name} [get]
//	@Summary	get a pipeline by name
//	@Tags		pipelines
//	@Produce	json
//	@Param		name	path		string	true	"pipeline name"
//	@Success	200		{object}	store.PipelineListItem
//	@Response	default	{object}	errorMsg "Error"
func (api *WebAPI) GetPipeline(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	d, err := api.q.PipelineGet(r.Context(), name)
	api.respond(w, r, d, err)
}

// GetPipelineList godoc
//
//	@Router		/pipelines [get]
//	@Summary	get a list of pipelines
//	@Tags		pipelines
//	@Produce	json
//	@Success	200		{array}		store.PipelineListItem
//	@Response	default	{object}	errorMsg "Error"
func (api *WebAPI) GetPipelineList(w http.ResponseWriter, r *http.Request) {
	d, err := api.q.PipelineList(r.Context())
	api.respond(w, r, d, err)
}

// GetPipelineRun godoc
//
//	@Router		/pipelines/{name}/runs [get]
//	@Summary	get runs for a pipeline
//	@Tags		pipelines
//	@Produce	json
//	@Param		name	path		string	true	"pipeline name"
//	@Param		status	query		string	false	"run status"
//	@Param		limit	query		int		false	"limit"
//	@Param		offset	query		int		false	"offset"
//	@Success	200		{array}		store.PipelineRunListRow
//	@Response	default	{object}	errorMsg "Error"
func (api *WebAPI) GetPipelineRunList(w http.ResponseWriter, r *http.Request) {
	var params store.PipelineRunListParams

	if err := decoder.Decode(&params, r.URL.Query()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if params.Limit == 0 {
		params.Limit = 100
	}

	if params.Status == nil {
		params.Status = stati
	}

	params.Name = mux.Vars(r)["name"]

	d, err := api.q.PipelineRunList(r.Context(), params)
	api.respond(w, r, d, err)
}

// GetPostHook godoc
//
//	@Router		/posthooks/{name} [get]
//	@Summary	get a posthook by name
//	@Tags		posthooks
//	@Produce	json
//	@Param		name	path		string	true	"posthook name"
//	@Success	200		{object}	store.PostHook
//	@Response	default	{object}	errorMsg "Error"
func (api *WebAPI) GetPostHook(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	d, err := api.q.PostHookGet(r.Context(), name)
	api.respond(w, r, d, err)
}

// GetPostHookList godoc
//
//	@Router		/posthooks [get]
//	@Summary	get a list of posthooks
//	@Tags		posthooks
//	@Produce	json
//	@Success	200		{array}		store.PostHook
//	@Response	default	{object}	errorMsg "Error"
func (api *WebAPI) GetPostHookList(w http.ResponseWriter, r *http.Request) {
	d, err := api.q.PostHookList(r.Context())
	api.respond(w, r, d, err)
}

// GetTask godoc
//
//	@Router		/tasks/{id} [get]
//	@Summary	get a task by name
//	@Tags		tasks
//	@Produce	json
//	@Param		id		path		string	true	"task id"
//	@Success	200		{object}	store.Task
//	@Response	default	{object}	errorMsg "Error"
func (a *WebAPI) GetTask(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	d, err := a.q.TaskGet(r.Context(), name)
	a.respond(w, r, d, err)
}

// GetTaskList godoc
//
//	@Router		/tasks [get]
//	@Summary	get a list of tasks
//	@Tags		tasks
//	@Produce	json
//	@Param		status	query		string	false	"task status"
//	@Param		limit	query		int		false	"limit"
//	@Param		offset	query		int		false	"offset"
//	@Success	200		{array}		store.Task
//	@Response	default	{object}	errorMsg "Error"
func (a *WebAPI) GetTaskList(w http.ResponseWriter, r *http.Request) {
	var params store.TaskListParams

	if err := decoder.Decode(&params, r.URL.Query()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if params.Limit == 0 {
		params.Limit = 100
	}

	if params.Status == nil {
		params.Status = stati
	}

	d, err := a.q.TaskList(r.Context(), params)
	a.respond(w, r, d, err)
}

// GetRun godoc
//
//	@Router		/runs/{id} [get]
//	@Summary	get a run by fingerprint
//	@Tags		runs
//	@Produce	json
//	@Param		id		path		string	true	"run fingerprint"
//	@Success	200		{object}	store.Run
//	@Response	default	{object}	errorMsg "Error"
func (api *WebAPI) GetRun(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	d, err := api.q.RunGet(r.Context(), name)
	api.respond(w, r, d, err)
}

// GetRunList godoc
//
//	@Router		/runs [get]
//	@Summary	get a list of runs
//	@Tags		runs
//	@Produce	json
//	@Param		status	query		string	false	"run status"
//	@Param		limit	query		int		false	"limit"
//	@Param		offset	query		int		false	"offset"
//	@Success	200		{array}		store.Run
//	@Response	default	{object}	errorMsg "Error"
func (api *WebAPI) GetRunList(w http.ResponseWriter, r *http.Request) {
	var params store.RunListParams

	if err := decoder.Decode(&params, r.URL.Query()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if params.Limit == 0 {
		params.Limit = 100
	}

	if params.Status == nil {
		params.Status = stati
	}

	d, err := api.q.RunList(r.Context(), params)
	api.respond(w, r, d, err)
}
