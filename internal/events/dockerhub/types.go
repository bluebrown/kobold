package dockerhub

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/bluebrown/kobold/internal/events"
	"github.com/qri-io/jsonschema"
)

var (
	//go:embed schema.json
	schemaBytes []byte
	schema      = &jsonschema.Schema{}
)

func init() {
	if err := json.Unmarshal(schemaBytes, schema); err != nil {
		panic("unmarshal schema: " + err.Error())
	}
}

type digestFetcher interface {
	Fetch(ref string) (digest string, err error)
}

type payloadHandler struct {
	schema        *jsonschema.Schema
	digestFetcher digestFetcher
}

func NewPayloadHandler(df digestFetcher) events.PayloadHandler {
	if df == nil {
		panic("digest fetcher is nil")
	}
	return payloadHandler{schema: schema, digestFetcher: df}
}

func (ph payloadHandler) Validate(b []byte, ct string) error {
	verr, err := ph.schema.ValidateBytes(context.TODO(), b)
	if err != nil {
		return err
	}
	if len(verr) > 0 {
		return fmt.Errorf("invalid data")
	}
	return nil
}

func (ph payloadHandler) Decode(b []byte, ct string) (events.PushData, error) {
	pl := PushPayload{}
	if err := json.Unmarshal(b, &pl); err != nil {
		return events.PushData{}, err
	}
	digest, err := ph.digestFetcher.Fetch(fmt.Sprintf("index.docker.io/%s:%s", pl.Repository.RepoName, pl.PushData.Tag))
	if err != nil {
		return events.PushData{}, err
	}
	return events.PushData{
		Image:  "index.docker.io/" + pl.Repository.RepoName,
		Tag:    pl.PushData.Tag,
		Digest: digest,
	}, nil
}

type PushPayload struct {
	CallbackURL string     `json:"callback_url"`
	PushData    PushData   `json:"push_data"`
	Repository  Repository `json:"repository"`
}

type PushData struct {
	Pusher    string        `json:"pusher"`
	PushedAt  int           `json:"pushed_at"`
	Tag       string        `json:"tag"`
	Images    []interface{} `json:"images"`
	MediaType string        `json:"media_type"`
}
type Repository struct {
	Status          string      `json:"status"`
	Namespace       string      `json:"namespace"`
	Name            string      `json:"name"`
	RepoName        string      `json:"repo_name"`
	RepoURL         string      `json:"repo_url"`
	Description     string      `json:"description"`
	FullDescription interface{} `json:"full_description"`
	StarCount       int         `json:"star_count"`
	IsPrivate       bool        `json:"is_private"`
	IsTrusted       bool        `json:"is_trusted"`
	IsOfficial      bool        `json:"is_official"`
	Owner           string      `json:"owner"`
	DateCreated     int         `json:"date_created"`
}
