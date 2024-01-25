package krm

import (
	"reflect"
	"testing"
)

func TestParseOpts(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
	t.Parallel()
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
