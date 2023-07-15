//go:build generate

package tools

import (
	_ "github.com/go-task/task/v3/cmd/task"
	_ "github.com/google/go-licenses"
)

//go:generate go run github.com/google/go-licenses check --include_tests ./cmd/server/ --allowed_licenses=0BSD,ISC,BSD-2-Clause,BSD-3-Clause,MIT,Apache-2.0
