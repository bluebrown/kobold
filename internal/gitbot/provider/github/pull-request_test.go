package github

import (
	"testing"
)

func TestUrlParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		giveUrl   string
		wantOwner string
		wantRepo  string
		wantError bool
	}{
		{
			name:      "https",
			giveUrl:   "https://github.com/bluebrown/kobold-test",
			wantOwner: "bluebrown",
			wantRepo:  "kobold-test",
			wantError: false,
		},
		{
			name:      "ssh",
			giveUrl:   "git@github.com:bluebrown/kobold-test.git",
			wantOwner: "bluebrown",
			wantRepo:  "kobold-test",
			wantError: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		owner, repo, err := getOwnerRepo(tt.giveUrl)

		if tt.wantError && err == nil {
			t.Errorf("%s: want error but got none", tt.name)
		}

		if owner != tt.wantOwner {
			t.Errorf("%s: want org %q but got %q", tt.name, tt.wantOwner, owner)
		}

		if repo != tt.wantRepo {
			t.Errorf("%s: want repo %q but got %q", tt.name, tt.wantRepo, repo)
		}

	}
}
