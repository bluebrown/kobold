package acr

import "testing"

func TestGoodSchema(t *testing.T) {
	ph := NewPayloadHandler()
	err := ph.Validate(goodPayload)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBadSchema(t *testing.T) {
	ph := NewPayloadHandler()
	err := ph.Validate(badPayload)
	if err == nil {
		t.Fatal("expected error but got none")
	}
}

var goodPayload = []byte(`{
	"action": "push",
	"id": "cb8c3971-9adc-488b-xxxx-43cbb4974ff5",
	"request": {
	  "host": "test.azurecr.io",
	  "id": "3cbb6949-7549-4fa1-xxxx-a6d5451dffc7",
	  "method": "PUT",
	  "useragent": "docker/17.09.0-ce go/go1.8.3 git-commit/afdb6d4 kernel/4.10.0-27-generic os/linux arch/amd64 UpstreamClient(Docker-Client/17.09.0-ce \\(linux\\))"
	},
	"target": {
	  "digest": "sha256:xxxxd5c8786bb9e621a45ece0dbxxxx1cdc624ad20da9fe62e9d25490f33xxxx",
	  "length": 524,
	  "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
	  "repository": "nginx",
	  "size": 524,
	  "tag": "v1"
	},
	"timestamp": "2017-11-17T16:52:01.343145347Z"
}`)

var badPayload = []byte(`{
	"action": "delete",
	"id": "cb8c3971-9adc-488b-xxxx-43cbb4974ff5",
	"request": {
	  "id": "3cbb6949-7549-4fa1-xxxx-a6d5451dffc7",
	  "method": "PUT",
	  "useragent": "docker/17.09.0-ce go/go1.8.3 git-commit/afdb6d4 kernel/4.10.0-27-generic os/linux arch/amd64 UpstreamClient(Docker-Client/17.09.0-ce \\(linux\\))"
	},
	"target": {
	  "length": 524,
	  "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
	  "repository": "nginx",
	  "size": 524
	},
	"timestamp": "2017-11-17T16:52:01.343145347Z"
}`)
