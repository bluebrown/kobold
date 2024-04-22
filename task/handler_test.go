package task

import (
	"testing"

	"github.com/bluebrown/kobold/krm"
)

func TestGetCommitMessage(t *testing.T) {
	type args struct {
		changes []krm.Change
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "single image",
			args: args{
				changes: []krm.Change{
					{
						Description: "busybox:1.0.0 -> busybox:1.0.1",
						Repo:        "busybox",
					},
				},
			},
			want:    "chore(kobold): Update busybox",
			wantErr: false,
		},
		{
			name: "multiple images",
			args: args{
				changes: []krm.Change{
					{
						Description: "busybox:1.0.0 -> busybox:1.0.1",
						Repo:        "busybox",
					},
					{
						Description: "somerepo:2.0.0 -> somerepo:2.0.1",
						Repo:        "somerepo",
					},
				},
			},
			want:    "chore(kobold): Update busybox somerepo",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCommitMessage(tt.args.changes)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommitMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetCommitMessage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
