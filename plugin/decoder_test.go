package plugin

import (
	"os"
	"testing"

	"github.com/bluebrown/kobold/krm"
	"github.com/bluebrown/kobold/plugin/builtin"
)

func TestDecoder(t *testing.T) {
	testCases := []struct {
		name       string
		decoder    string
		giveFile   string
		wantName   string
		wantTag    string
		wantDigest string
	}{
		{
			name:       "lines",
			decoder:    "decoder.lines@v1",
			giveFile:   "testdata/lines.txt",
			wantName:   "index.docker.io/bluebrown/busybox",
			wantTag:    "v1.1",
			wantDigest: "sha256:3b3128d9df6bbbcc92e2358e596c9fbd722a437a62bafbc51607970e9e3b8869",
		},
		{
			name:       "distribution",
			decoder:    "decoder.distribution@v1",
			giveFile:   "testdata/distribution.json",
			wantName:   "test.azurecr.io/busybox",
			wantTag:    "v1",
			wantDigest: "sha256:xxxxd5c8786bb9e621a45ece0dbxxxx1cdc624ad20da9fe62e9d25490f33xxxx",
		},
		{
			name:     "dockerhub",
			decoder:  "decoder.dockerhub@v1",
			giveFile: "testdata/dockerhub.json",
			wantName: "index.docker.io/svendowideit/testhook",
			wantTag:  "stable",
		},
		{
			name:       "harbor",
			decoder:    "decoder.harbor@v1",
			giveFile:   "testdata/harbor.json",
			wantName:   "ghcr.io/bluebrown/busybox",
			wantTag:    "v1.4",
			wantDigest: "sha256:3b3128d9df6bbbcc92e2358e596c9fbd722a437a62bafbc51607970e9e3b8869",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			dec := NewDecoderRunner()

			sb, err := builtin.StarlarkScripts.ReadFile("starlark/" + tc.decoder + ".star")
			if err != nil {
				t.Fatal(err)
			}

			fb, err := os.ReadFile(tc.giveFile)
			if err != nil {
				t.Fatal(err)
			}

			refs, err := dec.Decode(tc.decoder, sb, fb)
			if err != nil {
				t.Fatal(err)
			}

			if len(refs) != 1 {
				t.Errorf("ref mismatch, got %v, want %v", len(refs), 1)
			}

			tag, digest, err := krm.ParseImageRefWithDigest(refs[0])
			if err != nil {
				t.Fatalf("failed to parse image ref: %v", err)
			}

			if tag.Context().String() != tc.wantName {
				t.Errorf("got %v, want %v", tag.Context().String(), tc.wantName)
			}

			if tag.Identifier() != tc.wantTag {
				t.Errorf("got %v, want %v", tag.Identifier(), tc.wantTag)
			}

			if digest != tc.wantDigest {
				t.Errorf("got %v, want %v", digest, tc.wantDigest)
			}
		})
	}
}
