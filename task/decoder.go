package task

import (
	"fmt"

	"go.starlark.net/starlark"

	"github.com/bluebrown/kobold/starutil"
)

type StarlarkDecoder struct {
}

var _ Decoder = (*StarlarkDecoder)(nil)

func NewStarlarkDecoder() *StarlarkDecoder {
	return &StarlarkDecoder{}
}

func (d *StarlarkDecoder) Decode(name string, script []byte, data []byte) ([]string, error) {
	res, err := starutil.RunMain(starutil.DefaultThread(name), name, script, d.args(data), nil)
	if err != nil {
		return nil, fmt.Errorf("run main: %w", err)
	}
	return starutil.AsStringSlice(res)
}

func (d *StarlarkDecoder) args(data []byte) starlark.Tuple {
	return starlark.Tuple{starlark.String(data)}
}
