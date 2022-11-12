package generic

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
	"registry": "test.azurecr.io",
	"repository": "foo/nginx",
	"tag": "v1.1",
	"digest": "sha256:xxxxd5c8786bb9e621a45ece0dbxxxx1cdc624ad20da9fe62e9d25490f33xxxx"
}`)

var badPayload = []byte(`{
	"registry": "test.azurecr.io",
	"repository": "foo/nginx",
	"tag": "v1.1"
}`)
