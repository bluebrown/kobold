package plugin

import (
	"fmt"

	"go.starlark.net/starlark"
)

type Decoder struct{}

func NewDecoderRunner() *Decoder {
	return &Decoder{}
}

func (d *Decoder) Decode(name string, script []byte, data []byte) ([]string, error) {
	res, err := runMain(defaultThread(name), name, script, d.args(data), nil)
	if err != nil {
		return nil, fmt.Errorf("run main: %w", err)
	}
	return asStringSlice(res)
}

func (d *Decoder) args(data []byte) starlark.Tuple {
	return starlark.Tuple{starlark.String(data)}
}
