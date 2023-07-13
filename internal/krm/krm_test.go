package krm

import (
	"fmt"
	"testing"

	"github.com/bluebrown/kobold/internal/events"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestImageNodeHandler(t *testing.T) {
	tests := []struct {
		name          string
		giveImageRef  string
		giveOpts      string
		givePushData  events.PushData
		wantHasChange bool
		wantChange    Change
	}{
		{
			name:         "simple",
			giveImageRef: "bluebrown/busybox",
			giveOpts:     "tag: latest; type: exact",
			givePushData: events.PushData{
				Image:  "index.docker.io/bluebrown/busybox",
				Tag:    "latest",
				Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			},
			wantHasChange: true,
			wantChange: Change{
				OldImageRef: "bluebrown/busybox",
				NewImageRef: "index.docker.io/bluebrown/busybox:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			},
		},
		{
			name:          "tag",
			giveImageRef:  "bluebrown/busybox:latest",
			giveOpts:      "tag: latest; type: exact",
			wantHasChange: true,
			givePushData: events.PushData{
				Image:  "index.docker.io/bluebrown/busybox",
				Tag:    "latest",
				Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			},
			wantChange: Change{
				OldImageRef: "bluebrown/busybox:latest",
				NewImageRef: "index.docker.io/bluebrown/busybox:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			},
		},
		{
			name:         "digest",
			giveImageRef: "bluebrown/busybox@sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3",
			giveOpts:     "tag: latest; type: exact",
			givePushData: events.PushData{
				Image:  "index.docker.io/bluebrown/busybox",
				Tag:    "latest",
				Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			},
			wantHasChange: true,
			wantChange: Change{
				OldImageRef: "bluebrown/busybox@sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3",
				NewImageRef: "index.docker.io/bluebrown/busybox:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			},
		},
		{
			name:         "tag+digest",
			giveImageRef: "bluebrown/busybox:latest@sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3",
			giveOpts:     "tag: latest; type: exact",
			givePushData: events.PushData{
				Image:  "index.docker.io/bluebrown/busybox",
				Tag:    "latest",
				Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			},
			wantHasChange: true,
			wantChange: Change{
				OldImageRef: "bluebrown/busybox:latest@sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3",
				NewImageRef: "index.docker.io/bluebrown/busybox:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			},
		},
		{
			name:         "no change with tag+digest",
			giveImageRef: "index.docker.io/bluebrown/busybox:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			giveOpts:     "tag: latest; type: exact",
			givePushData: events.PushData{
				Image:  "index.docker.io/bluebrown/busybox",
				Tag:    "latest",
				Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			},
			wantHasChange: false,
			wantChange:    Change{},
		},
		{
			name:         "regex semver",
			giveImageRef: "test.azurecr.io/nginx:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			giveOpts:     "tag: v[0-9]+.[0-9]+.[0-9]+; type: regex",
			givePushData: events.PushData{
				Image:  "test.azurecr.io/nginx",
				Tag:    "v2.1.0",
				Digest: "sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3",
			},
			wantHasChange: true,
			wantChange: Change{
				OldImageRef: "test.azurecr.io/nginx:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
				NewImageRef: "test.azurecr.io/nginx:v2.1.0@sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3",
			},
		},
		{
			name:         "regex semver not beta",
			giveImageRef: "test.azurecr.io/nginx:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			giveOpts:     "tag: v[0-9]+.[0-9]+.[0-9]+; type: regex",
			givePushData: events.PushData{
				Image:  "test.azurecr.io/nginx",
				Tag:    "v2.1.0-beta.1",
				Digest: "sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3",
			},
			wantHasChange: false,
			wantChange:    Change{},
		},
		{
			name:         "regex semver only beta",
			giveImageRef: "test.azurecr.io/nginx:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			giveOpts:     `tag: v[0-9]+.[0-9]+.[0-9]+-beta.[0-9]+; type: regex`,
			givePushData: events.PushData{
				Image:  "test.azurecr.io/nginx",
				Tag:    "v2.1.0-beta.1",
				Digest: "sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3",
			},
			wantHasChange: true,
			wantChange: Change{
				OldImageRef: "test.azurecr.io/nginx:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
				NewImageRef: "test.azurecr.io/nginx:v2.1.0-beta.1@sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := NewImageNodeHandler(DefaultImagerefTemplate)
			rnode := yaml.MustParse(fmt.Sprintf("image: %s %s%s", tt.giveImageRef, CommentPrefix, tt.giveOpts))
			hasChange, change, err := handler.HandleImageNode(rnode.Field("image"), []events.PushData{tt.givePushData})
			if err != nil {
				t.Fatal(err)
			}
			if hasChange != tt.wantHasChange {
				t.Errorf("wrong has change value: got %v but want %v", hasChange, tt.wantHasChange)
			}
			if change.OldImageRef != tt.wantChange.OldImageRef {
				t.Errorf("wrong OldImageRef: got %q but want %q", change.OldImageRef, tt.wantChange.OldImageRef)
			}
			if change.NewImageRef != tt.wantChange.NewImageRef {
				t.Errorf("wrong NewImageRef: got %q but want %q", change.NewImageRef, tt.wantChange.NewImageRef)
			}
		})
	}
}
