package dockerhub

import (
	"testing"
)

type nopFetcher struct{}

func (nopFetcher) Fetch(ref string) (string, error) {
	return "", nil
}

func TestGoodSchema(t *testing.T) {
	ph := NewPayloadHandler(nopFetcher{})
	err := ph.Validate(goodPayload)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBadSchema(t *testing.T) {
	ph := NewPayloadHandler(nopFetcher{})
	err := ph.Validate(badPayload)
	if err == nil {
		t.Fatal("expected error but got none")
	}
}

var goodPayload = []byte(`{
	"callback_url": "https://registry.hub.docker.com/u/foo/busybox/hook/565a7531f7ab4a1283624816984a5bc5/",
	"push_data": {
	  "pusher": "foo",
	  "pushed_at": 1671307562,
	  "tag": "v1.1",
	  "images": [],
	  "media_type": "application/vnd.docker.distribution.manifest.list.v2+json"
	},
	"repository": {
	  "status": "Active",
	  "namespace": "bluebrown",
	  "name": "busybox",
	  "repo_name": "foo/busybox",
	  "repo_url": "https://hub.docker.com/r/foo/busybox",
	  "description": "",
	  "full_description": null,
	  "star_count": 0,
	  "is_private": false,
	  "is_trusted": false,
	  "is_official": false,
	  "owner": "foo",
	  "date_created": 1671305618
	}
}`)

var badPayload = []byte(`{
	"callback_url": "https://registry.hub.docker.com/u/bluebrown/busybox/hook/565a7531f7ab4a1283624816984a5bc5/",
	"push_data": {
	  "pusher": "bluebrown",
	  "pushed_at": 1671307562,
	  "images": [],
	  "media_type": "application/vnd.docker.distribution.manifest.list.v2+json"
	},
	"repository": {
	  "status": "Active",
	  "namespace": "bluebrown",
	  "name": "busybox",
	  "description": "",
	  "full_description": null,
	  "star_count": 0,
	  "is_private": false,
	  "is_trusted": false,
	  "is_official": false,
	  "date_created": 1671305618
	}
}`)
