package builtin

import (
	"embed"
	"path/filepath"
	"strings"
)

//go:embed starlark/*
var StarlarkScripts embed.FS

func Decoders() []data {
	return read("decoder")
}

func PostHooks() []data {
	return read("posthook")
}

type data struct {
	Name   string
	Script string
}

func read(kind string) []data {
	dir, err := StarlarkScripts.ReadDir("starlark")
	if err != nil {
		panic(err)
	}

	var items []data
	for _, f := range dir {
		if f.IsDir() {
			continue
		}

		if !strings.HasPrefix(f.Name(), kind+".") {
			continue
		}

		name := strings.TrimPrefix(f.Name(), kind+".")
		name = strings.TrimSuffix(name, ".star")

		script, err := StarlarkScripts.ReadFile(filepath.Join("starlark", f.Name()))
		if err != nil {
			panic(err)
		}

		items = append(items, data{
			Name:   "builtin." + name,
			Script: string(script),
		})

	}

	return items
}
