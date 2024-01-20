package plugin

import (
	"reflect"
	"testing"

	"go.starlark.net/starlark"
)

func TestRunMain(t *testing.T) {
	type args struct {
		thread  *starlark.Thread
		name    string
		script  []byte
		args    starlark.Tuple
		hostEnv *starlark.Dict
	}
	tests := []struct {
		name    string
		args    args
		want    starlark.Value
		wantErr bool
	}{
		{
			name: "simple",
			args: args{
				thread: defaultThread("test"),
				name:   "test",
				script: []byte(`def main(): return 1`),
				args:   starlark.Tuple{},
			},
			want: starlark.MakeInt(1),
		},
		{
			name: "lookup env",
			args: args{
				thread: defaultThread("test"),
				name:   "test",
				script: []byte(`def main(): return host_env["FOO"]`),
				args:   starlark.Tuple{},
				hostEnv: func() *starlark.Dict {
					d := starlark.NewDict(0)
					if err := d.SetKey(starlark.String("FOO"), starlark.String("BAR")); err != nil {
						panic(err)
					}
					return d
				}(),
			},
			want: starlark.String("BAR"),
		},
		{
			name: "pass args",
			args: args{
				thread: defaultThread("test"),
				name:   "test",
				script: []byte(`def main(a, b): return a + b`),
				args:   starlark.Tuple{starlark.MakeInt(1), starlark.MakeInt(2)},
			},
			want: starlark.MakeInt(3),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runMain(tt.args.thread, tt.args.name, tt.args.script, tt.args.args, tt.args.hostEnv)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunMain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RunMain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvToStarlarkDict(t *testing.T) {
	tests := []struct {
		name     string
		giveEnv  []string
		wantDict *starlark.Dict
	}{
		{
			name:    "simple",
			giveEnv: []string{"FOO=BAR"},
			wantDict: func() *starlark.Dict {
				d := starlark.NewDict(0)
				if err := d.SetKey(starlark.String("FOO"), starlark.String("BAR")); err != nil {
					panic(err)
				}
				return d
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := envToStarlarkDict(tt.giveEnv); !reflect.DeepEqual(got, tt.wantDict) {
				t.Errorf("OsEnvToStarlarkDict() = %v, want %v", got, tt.wantDict)
			}
		})
	}
}

func TestAsStringSlice(t *testing.T) {
	type args struct {
		v starlark.Value
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{

		{
			name: "simple",
			args: args{v: starlark.NewList([]starlark.Value{starlark.String("foo"), starlark.String("bar")})},
			want: []string{"foo", "bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := asStringSlice(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("AsStringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AsStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
