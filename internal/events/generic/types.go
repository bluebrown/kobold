package generic

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

func (ph payloadHandler) Decode(b []byte) (events.PushData, error) {
	pl := PushPayload{}
	if err := json.Unmarshal(b, &pl); err != nil {
		return events.PushData{}, err
	}
	return events.PushData{
		Image:  fmt.Sprintf("%s/%s", pl.Registry, pl.Repository),
		Tag:    pl.ImageTag,
		Digest: pl.ImageDigest,
	}, nil
}

func init() {
	if err := json.Unmarshal(schemaBytes, schema); err != nil {
		panic("unmarshal schema: " + err.Error())
	}
}

type PushPayload struct {
	Registry    string `json:"registry,omitempty"`
	Repository  string `json:"repository,omitempty"`
	ImageTag    string `json:"tag,omitempty"`
	ImageDigest string `json:"digest,omitempty"`
}
