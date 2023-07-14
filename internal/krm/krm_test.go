package krm

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bluebrown/kobold/internal/events"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestParseOpts(t *testing.T) {
	type args struct {
		expr string
	}
	tests := []struct {
		name    string
		args    args
		want    Options
		wantErr bool
	}{
		{
			name: "Latest Exact",
			args: args{expr: "tag: latest; type: exact"},
			want: Options{Type: TypeExact, Tag: "latest"},
		},
		{
			name: "Short Semver",
			args: args{expr: "tag: ^1; type: semver"},
			want: Options{Type: TypeSemver, Tag: "^1"},
		},
		{
			name: "Comma Semver",
			args: args{expr: "tag: >= 2.3, < 3; type: semver"},
			want: Options{Type: TypeSemver, Tag: ">= 2.3, < 3"},
		},
		{
			name: "Regex",
			args: args{expr: "tag: foo-.*; type: regex"},
			want: Options{Type: TypeRegex, Tag: "foo-.*"},
		},
		{
			name:    "Unknown Key",
			args:    args{expr: "nope: true"},
			wantErr: true,
		},
		{
			name:    "Not Key Value Pair",
			args:    args{expr: "nope"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOpts(tt.args.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOpts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseOpts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchTag(t *testing.T) {
	type args struct {
		tag  string
		opts Options
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "exact match",
			args: args{tag: "latest", opts: Options{Type: TypeExact, Tag: "latest"}},
			want: true,
		},
		{
			name: "exact no-match",
			args: args{tag: "dev", opts: Options{Type: TypeExact, Tag: "other"}},
			want: false,
		},
		{
			name: "semver match",
			args: args{tag: "v0.2.13", opts: Options{Type: TypeSemver, Tag: "^0"}},
			want: true,
		},
		{
			name: "semver no-match",
			args: args{tag: "v1.5.67", opts: Options{Type: TypeSemver, Tag: "<=1.4"}},
			want: false,
		},
		{
			name:    "semver invalid",
			args:    args{tag: "v1.0.0", opts: Options{Type: TypeSemver, Tag: "kaboom"}},
			wantErr: true,
		},
		{
			name: "regex match",
			args: args{tag: "sprint_8675343", opts: Options{Type: TypeRegex, Tag: "sprint_\\d+"}},
			want: true,
		},
		{
			name:    "regex invalid",
			args:    args{tag: "sprint_8675343", opts: Options{Type: TypeRegex, Tag: "("}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchTag(tt.args.tag, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MatchTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ignore(t *testing.T) {
	type wants struct {
		relPath string
		ignore  bool
	}
	tests := []struct {
		name       string
		giveScopes []string
		wants      []wants
	}{
		{
			name:       "simple exact",
			giveScopes: []string{"stage/deployment.yaml"},
			wants: []wants{
				{
					relPath: "stage/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "prod/deployment.yaml",
					ignore:  true,
				},
			},
		},
		{
			name:       "top level dir",
			giveScopes: []string{"stage/"},
			wants: []wants{
				{
					relPath: "stage/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "prod/deployment.yaml",
					ignore:  true,
				},
			},
		},
		{
			name:       "nested level dir",
			giveScopes: []string{"stage/"},
			wants: []wants{
				{
					relPath: "my-app/stage/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "my-app/prod/deployment.yaml",
					ignore:  true,
				},
			},
		},
		{
			name:       "from top level only",
			giveScopes: []string{"/stage/"},
			wants: []wants{
				{
					relPath: "stage/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "my-app/stage/deployment.yaml",
					ignore:  true,
				},
			},
		},
		{
			name:       "mutli scope",
			giveScopes: []string{"stage/", "dev/"},
			wants: []wants{
				{
					relPath: "stage/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "my-app/stage/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "dev/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "my-app/dev/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "prod/deployment.yaml",
					ignore:  true,
				},
				{
					relPath: "my-app/prod/deployment.yaml",
					ignore:  true,
				},
			},
		},
		{
			name:       "file negation",
			giveScopes: []string{"*.yaml", "!compose.yaml"},
			wants: []wants{
				{
					relPath: "deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "compose.yaml",
					ignore:  true,
				},
			},
		},
		// FIXME: folder negation is not working
		// {
		// 	name:       "folder negation",
		// 	giveScopes: []string{"!/prod/"},
		// 	wants: []wants{
		// 		{
		// 			relPath: "prod/deployment.yaml",
		// 			ignore:  true,
		// 		},
		// 		{
		// 			relPath: "dev/deployment.yaml",
		// 			ignore:  false,
		// 		},
		// 		{
		// 			relPath: "stage/deployment.yaml",
		// 			ignore:  false,
		// 		},
		// 	},
		// },
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fn := ignore(tt.giveScopes)
			for _, w := range tt.wants {
				m := fn(w.relPath)
				if m != w.ignore {
					t.Errorf("%q did not have correct ignore. Want %v but got %v", w.relPath, w.ignore, m)
				}
			}
		})
	}
}

func TestImageNodeHandler_HandleImageNode(t *testing.T) {
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
