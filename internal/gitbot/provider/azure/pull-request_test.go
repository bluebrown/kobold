package azure

import (
	"testing"
)

func TestUrlParse(t *testing.T) {
	tests := []struct {
		name        string
		giveUrl     string
		wantOrg     string
		wantProject string
		wantRepoId  string
		wantError   bool
	}{
		{
			name:        "https",
			giveUrl:     "https://my-org@dev.azure.com/my-org/my-project/_git/my-repo",
			wantOrg:     "my-org",
			wantProject: "my-project",
			wantRepoId:  "my-repo",
			wantError:   false,
		},
		{
			name:        "https-git-suffix",
			giveUrl:     "https://my-org@dev.azure.com/my-org/my-project/_git/my-repo.git",
			wantOrg:     "my-org",
			wantProject: "my-project",
			wantRepoId:  "my-repo",
			wantError:   false,
		},
		{
			name:        "ssh",
			giveUrl:     "git@ssh.dev.azure.com:v3/my-org/my-project/my-repo",
			wantOrg:     "my-org",
			wantProject: "my-project",
			wantRepoId:  "my-repo",
			wantError:   false,
		},
		{
			name:        "ssh-git-suffix",
			giveUrl:     "git@ssh.dev.azure.com:v3/my-org/my-project/my-repo.git",
			wantOrg:     "my-org",
			wantProject: "my-project",
			wantRepoId:  "my-repo",
			wantError:   false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			org, proj, repo, err := getOrgProjRepo(tt.giveUrl)

			if tt.wantError && err == nil {
				t.Errorf("want error but got none")
			}

			if org != tt.wantOrg {
				t.Errorf("want org %q but got %q", tt.wantOrg, org)
			}

			if proj != tt.wantProject {
				t.Errorf("want proj %q but got %q", tt.wantProject, proj)
			}

			if repo != tt.wantRepoId {
				t.Errorf("want repo %q but got %q", tt.wantRepoId, repo)
			}
		})

	}
}
