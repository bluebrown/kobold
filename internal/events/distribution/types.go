package distribution

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

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

type payloadHandler struct {
	schema *jsonschema.Schema
}

func NewPayloadHandler() events.PayloadHandler {
	return payloadHandler{schema: schema}
}

func (ph payloadHandler) Validate(b []byte) error {
	verr, err := ph.schema.ValidateBytes(context.TODO(), b)
	if err != nil {
		return err
	}
	if len(verr) > 0 {
		return fmt.Errorf("invalid data")
	}
	return nil
}

func (ph payloadHandler) Decode(b []byte) ([]events.PushData, error) {
	ple := PushPayloadEnvelope{}
	if err := json.Unmarshal(b, &ple); err != nil {
		return nil, err
	}
	pushDataSlice := make([]events.PushData, len(ple.Events))
	for i, pl := range ple.Events {
		pushDataSlice[i] = events.PushData{
			Image:  fmt.Sprintf("%s/%s", pl.Request.Host, pl.Target.Repository),
			Tag:    pl.Target.Tag,
			Digest: pl.Target.Digest,
		}
	}
	return pushDataSlice, nil
}

type PushPayloadEnvelope struct {
	Events []PushPayload `json:"events"`
}

type PushPayload struct {
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Action    string      `json:"action"`
	Target    PushTarget  `json:"target"`
	Request   PushRequest `json:"request"`
}

type PushTarget struct {
	MediaType  string `json:"mediaType"`
	Size       int    `json:"size"`
	Digest     string `json:"digest"`
	Length     int    `json:"length"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

type PushRequest struct {
	ID        string `json:"id"`
	Host      string `json:"host"`
	Method    string `json:"method"`
	Useragent string `json:"useragent"`
}
