/* read a user friendly config file, and prepare the database accordingly. */
package config

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/bluebrown/kobold/kioutil"
)

func TestReadDir(t *testing.T) {
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
					RepoURI: kioutil.GitPackageURI{
						Repo: "git@github.com:bluebrown/testing",
						Ref:  "dev",
						Pkg:  "/resources",
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
		t.Run(tt.givePath, func(t *testing.T) {
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

				toml.NewEncoder(&b1).Encode(got)
				toml.NewEncoder(&b2).Encode(tt.wantConfig)

				t.Errorf("ReadFile() got:\n%s\n\nwant:\n%s\n", b1.String(), b2.String())

			}
		})
	}
}
