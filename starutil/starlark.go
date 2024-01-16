package starutil

// TODO: dont panic on error

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/qri-io/starlib"
	"go.starlark.net/starlark"
)

func RunMain(thread *starlark.Thread, name string, script []byte, args starlark.Tuple, hostEnv *starlark.Dict) (starlark.Value, error) {
	globals := starlark.StringDict{
		"host_env": hostEnv,
	}
	d, err := starlark.ExecFile(thread, name+".star", script, globals)
	if err != nil {
		return nil, fmt.Errorf("exec: %w", err)
	}
	m, ok := d["main"]
	if !ok {
		return nil, fmt.Errorf("no main function defined")
	}
	return starlark.Call(thread, m, args, nil)
}

func AsStringSlice(v starlark.Value) ([]string, error) {
	_, ok := v.(starlark.Iterable)
	if !ok {
		return nil, fmt.Errorf("expected iterable, got %s", v.Type())
	}
	iterator := starlark.Iterate(v)
	defer iterator.Done()
	var item starlark.Value
	var slice []string
	for iterator.Next(&item) {
		s, ok := starlark.AsString(item)
		if !ok {
			return nil, fmt.Errorf("expected string, got %s", item.Type())
		}
		slice = append(slice, s)
	}
	return slice, nil
}

func EnvToStarlarkDict(env []string) *starlark.Dict {
	d := starlark.NewDict(0)
	for _, e := range env {
		key, val, ok := strings.Cut(e, "=")
		if !ok {
			continue
		}
		if err := d.SetKey(starlark.String(key), starlark.String(val)); err != nil {
			panic(err)
		}
	}
	return d
}

func DefaultThread(name string) *starlark.Thread {
	return &starlark.Thread{
		Name: name,
		Load: starlib.Loader,
		Print: func(thread *starlark.Thread, msg string) {
			slog.Info("post hook", "msg", msg, "fingerprint", thread.Name)
		},
	}
}
