package distribution

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
	"events": [
		{
			"action": "push",
			"id": "6e401fa1-c6d8-48ae-bbfe-2d06941f11b6",
			"request": {
				"host": "test.azurecr.io",
				"id": "d89e7b46-38b0-4f71-83d3-ba6fc005c189",
				"method": "PUT",
				"useragent": "docker/24.0.6 go/go1.20.7 git-commit/1a79695 kernel/5.10.0-25-amd64 os/linux arch/amd64 UpstreamClient(Docker-Client/24.0.6 \\(linux\\))"
			},
			"target": {
				"digest": "sha256:xxxxd5c8786bb9e621a45ece0dbxxxx1cdc624ad20da9fe62e9d25490f33xxxx",
				"length": 524,
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"repository": "nginx",
				"size": 524,
				"tag": "v1"
			},
			"timestamp": "2023-10-11T14:45:59.730519823Z"
		},
		{
			"action": "push",
			"id": "9a5f8995-dbab-4c2b-b06c-0c29c23e759e",
			"request": {
				"host": "test.azurecr.io",
				"id": "ae93b7e8-7e96-4e21-8487-917d74224d92",
				"method": "PUT",
				"useragent": "docker/24.0.6 go/go1.20.7 git-commit/1a79695 kernel/5.10.0-25-amd64 os/linux arch/amd64 UpstreamClient(Docker-Client/24.0.6 \\(linux\\))"
			},
			"target": {
				"digest": "sha256:xxxxd5c8786bb9e621a45ece0dbxxxx1cdc624ad20da9fe62e9d25490f33xxxx",
				"length": 524,
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"repository": "busybox",
				"size": 524,
				"tag": "v1"
			},
			"timestamp": "2023-10-11T14:45:59.730519823Z"
		}
	]
}`)

var badPayload = []byte(`{
	"events": [
		{
			"action": "push",
			"id": "2f21f8f8-431f-49f9-8f4e-c080094dfc71",
			"request": {
				"host": "test.azurecr.io",
				"id": "407cb475-dadb-4bd1-995f-15eddbdc98e2",
				"method": "PUT",
				"useragent": "docker/24.0.6 go/go1.20.7 git-commit/1a79695 kernel/5.10.0-25-amd64 os/linux arch/amd64 UpstreamClient(Docker-Client/24.0.6 \\(linux\\))"
			},
			"target": {
				"digest": "sha256:xxxxd5c8786bb9e621a45ece0dbxxxx1cdc624ad20da9fe62e9d25490f33xxxx",
				"length": 524,
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"repository": "nginx",
				"size": 524,
				"tag": "v1"
			},
			"timestamp": "2023-10-11T14:45:59.730519823Z"
		},
		{
			"action": "delete",
			"id": "71695e9f-988a-4992-a49f-c52904f7abf0",
			"request": {
				"host": "test.azurecr.io",
				"id": "341ee113-8ea2-4138-9d0d-8e0eddadd55a",
				"method": "PUT",
				"useragent": "docker/24.0.6 go/go1.20.7 git-commit/1a79695 kernel/5.10.0-25-amd64 os/linux arch/amd64 UpstreamClient(Docker-Client/24.0.6 \\(linux\\))"
			},
			"target": {
				"digest": "sha256:xxxxd5c8786bb9e621a45ece0dbxxxx1cdc624ad20da9fe62e9d25490f33xxxx",
				"length": 524,
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"repository": "busybox",
				"size": 524,
				"tag": "v1"
			},
			"timestamp": "2023-10-11T14:45:59.730519823Z"
		}
	]
}`)
