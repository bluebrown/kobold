package plugin

import (
	"fmt"

	"go.starlark.net/starlark"
)

type decoder struct {
}

func NewDecoderRunner() *decoder {
	return &decoder{}
}

func (d *decoder) Decode(name string, script []byte, data []byte) ([]string, error) {
	res, err := runMain(defaultThread(name), name, script, d.args(data), nil)
	if err != nil {
		return nil, fmt.Errorf("run main: %w", err)
	}
	return asStringSlice(res)
}

func (d *decoder) args(data []byte) starlark.Tuple {
	return starlark.Tuple{starlark.String(data)}
}
