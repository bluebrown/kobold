package generic

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bluebrown/kobold/internal/events"
	"github.com/google/go-containerregistry/pkg/name"
)

type payloadHandler struct {
}

func NewPayloadHandler() events.PayloadHandler {
	return payloadHandler{}
}

func (ph payloadHandler) Validate(b []byte) error {
	_, err := ph.Decode(b)
	return err
}

func (ph payloadHandler) Decode(b []byte) (events.PushData, error) {
	rawRef, digest, found := strings.Cut(string(b), "@")
	if !found {
		return events.PushData{}, errors.New("missing digest")
	}

	tag, err := name.NewTag(rawRef, name.StrictValidation)
	if err != nil {
		return events.PushData{}, err
	}

	return events.PushData{
		Image:  fmt.Sprintf("%s/%s", tag.RegistryStr(), tag.RepositoryStr()),
		Tag:    tag.TagStr(),
		Digest: digest,
	}, nil
}
