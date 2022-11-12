package registry

import (
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
)

func TestDigest(t *testing.T) {
	d, err := fetchDigest("index.docker.io/bluebrown/echoserver:latest", name.DefaultRegistry, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(d)
}
