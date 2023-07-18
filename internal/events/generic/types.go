package generic

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/bluebrown/kobold/internal/events"
	//se this parser instead as the google one dosnt give acces to tag(private variable)
	docker "github.com/novln/docker-parser"
	"github.com/qri-io/jsonschema"
	"strings"
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

func (ph payloadHandler) Validate(b []byte, ct string) error {
	switch ct {
	case "application/json":
		verr, err := ph.schema.ValidateBytes(context.TODO(), b)
		if err != nil {
			return err
		}
		if len(verr) > 0 {
			return fmt.Errorf("invalid data")
		}
		return nil
	case "text/plain":
		var text = string(b[:])
		if len(text) < 10 {
			return fmt.Errorf("invalid data")
		}
		return nil
	default:
		return fmt.Errorf("invalid Content-Type: %s", ct)
	}

}

func (ph payloadHandler) Decode(b []byte, ct string) (events.PushData, error) {
	pl := PushPayload{}
	switch ct {
	case "application/json":
		if err := json.Unmarshal(b, &pl); err != nil {
			return events.PushData{}, err
		}
		return events.PushData{
			Image:  fmt.Sprintf("%s/%s", pl.Registry, pl.Repository),
			Tag:    pl.ImageTag,
			Digest: pl.ImageDigest,
		}, nil
	case "text/plain":
		leftString, digest, _ := strings.Cut(string(b[:]), "@")

		newRef2, err := docker.Parse(leftString)

		if err != nil {
			return events.PushData{}, err
		}
		return events.PushData{
			Image:  newRef2.Repository(),
			Tag:    newRef2.Tag(),
			Digest: digest,
		}, nil

	default:
		return events.PushData{}, fmt.Errorf("invalid Content-Type: %s", ct)
	}
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
