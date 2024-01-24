package config

import "testing"

func TestLookup(t *testing.T) {
	t.Parallel()
	type args struct {
		env      []string
		key      string
		fallback string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "fallback",
			args: args{
				env:      []string{},
				key:      "KOBOLD_SQLITE_FILE",
				fallback: "kobold.sqlite3",
			},
			want: "kobold.sqlite3",
		},
		{
			name: "env",
			args: args{
				env:      []string{"KOBOLD_SQLITE_FILE=foo.sqlite3"},
				key:      "KOBOLD_SQLITE_FILE",
				fallback: "kobold.sqlite3",
			},
			want: "foo.sqlite3",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got, _ := Lookup(tt.args.env, tt.args.key, tt.args.fallback); got != tt.want {
				t.Errorf("lookup() = %v, want %v", got, tt.want)
			}
		})
	}
}
