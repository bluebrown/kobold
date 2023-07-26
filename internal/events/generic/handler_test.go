package generic

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/bluebrown/kobold/internal/events"
)

func Test_payloadHandler_Validate(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "valid ref",
			args:    args{b: []byte(`index.docker.io/foo/bar:v1@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248`)},
			wantErr: false,
		},
		{
			name:    "no registry",
			args:    args{b: []byte(`yum/yam:latest@sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3`)},
			wantErr: true,
		},
		{
			name:    "no tag",
			args:    args{b: []byte(`azurecr.io/some/thing@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248`)},
			wantErr: true,
		},
		{
			name:    "no digest",
			args:    args{b: []byte(`index.docker.io/library/busybox:123`)},
			wantErr: true,
		},
		{
			name:    "no namespace",
			args:    args{b: []byte(`index.docker.io/redis:7.4@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248`)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ph := NewPayloadHandler()
			if err := ph.Validate(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("payloadHandler.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_payloadHandler_Decode(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    events.PushData
		wantErr bool
	}{
		{
			name: "simple",
			args: args{b: []byte(`index.docker.io/foo/bar:v1@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248`)},
			want: events.PushData{
				Image:  "index.docker.io/foo/bar",
				Tag:    "v1",
				Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ph := NewPayloadHandler()
			got, err := ph.Decode(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("payloadHandler.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("payloadHandler.Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}
