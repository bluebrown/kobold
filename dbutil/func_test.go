package dbutil

import "testing"

func TestTaskFingerPrint(t *testing.T) {
	type args struct {
		ids []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "single",
			args: args{ids: []string{
				"d8fe8dad-c109-4f50-b37c-45bc9d063bdc",
			}},
			want: "37a98e8df308676a05105f70dfe0dc7e65422e0b",
		},
		{
			name: "multiple",
			args: args{ids: []string{
				"9b8cd0b7-0537-41cb-94b7-eca4651f6ece",
				"fc95ce25-880d-4ccd-aa74-0a644e2d9adf",
				"a97de96e-4452-4aeb-81cc-f87c4ec45d21",
			}},
			want: "804c20181f33e33990ae4769fa5ce9f2dc36e0ef",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TaskFingerPrint(tt.args.ids); got != tt.want {
				t.Errorf("TaskFingerPrint() = %v, want %v", got, tt.want)
			}
		})
	}
}
