package generic

import (
	"fmt"
	"testing"
)

func TestGoodSchema(t *testing.T) {
	ph := NewPayloadHandler()
	err := ph.Validate(goodJsonPayload, "application/json")
	if err != nil {
		t.Fatal(err)
	}
}
func TestBadContentTypeDecode(t *testing.T) {
	ph := NewPayloadHandler()
	var ct = "SomeBadContentType"
	_, err := ph.Decode(goodJsonPayload, ct)

	if err == nil {
		t.Fatal("expected error but got none")
	}
	t.Log(err.Error())
	if err == fmt.Errorf("invalid Content-Type: %s", ct) {
		t.Fatal("expected invalid Content-Type error but got none")
	}
}
func TestGoodObjectContentTypeDecode(t *testing.T) {
	ph := NewPayloadHandler()
	var ct = "application/json"
	_, err := ph.Decode(goodJsonPayload, ct)

	if err != nil {
		t.Fatal(err)
	}

}
func TestGoodStringContentTypeDecode(t *testing.T) {
	ph := NewPayloadHandler()
	var ct = "string"
	_, err := ph.Decode(goodStringPayload, ct)

	if err != nil {
		t.Fatal(err)
	}
}
func TestBadSchema(t *testing.T) {
	ph := NewPayloadHandler()
	err := ph.Validate(badPayload, "application/json")
	if err == nil {
		t.Fatal("expected error but got none")
	}

}

var goodJsonPayload = []byte(`{
	"registry": "test.azurecr.io",
	"repository": "foo/nginx",
	"tag": "v1.1",
	"digest": "sha256:xxxxd5c8786bb9e621a45ece0dbxxxx1cdc624ad20da9fe62e9d25490f33xxxx"
}`)
var goodStringPayload = []byte("test.azurecr.io/foo/nginx:v1.1@sha256:xxxxd5c8786bb9e621a45ece0dbxxxx1cdc624ad20da9fe62e9d25490f33xxxx")

var badPayload = []byte(`{
	"registry": "test.azurecr.io",
	"repository": "foo/nginx",
	"tag": "v1.1"
}`)
