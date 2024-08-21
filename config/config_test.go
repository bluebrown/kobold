/* Read a user friendly config file, and prepare the database accordingly. */
package config

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/bluebrown/kobold/git"
)

func TestReadDir(t *testing.T) {
	t.Parallel()
	tests := []struct {
		givePath   string
		wantConfig *Config
		wantErr    bool
	}{
		{
			givePath: "split",
			wantConfig: &Config{
				Version:  "2",
				Channels: []Channel{{Name: "foo"}},
				Pipelines: []Pipeline{{
					Name: "bar",
					RepoURI: git.PackageURI{
						Repo: "git@github.com:bluebrown/testing",
						Ref:  "dev",
						Pkg:  "resources",
					},
					Channels: []string{"foo"},
				}},
			},
		},
		{
			givePath: "symlinks",
			wantConfig: &Config{
				Version:  "2",
				Channels: []Channel{{Name: "symlinks"}},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.givePath, func(t *testing.T) {
			t.Parallel()
			var got Config
			err := ReadConfD(filepath.Join("testdata", tt.givePath), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(&got, tt.wantConfig) {
				var (
					b1 bytes.Buffer
					b2 bytes.Buffer
				)

				if err := toml.NewEncoder(&b1).Encode(got); err != nil {
					t.Fatal(err)
				}

				if err := toml.NewEncoder(&b2).Encode(tt.wantConfig); err != nil {
					t.Fatal(err)
				}

				t.Errorf("ReadFile() got:\n%s\n\nwant:\n%s\n", b1.String(), b2.String())
			}
		})
	}
}
