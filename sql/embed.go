package sql

import _ "embed"

//go:embed task.schema.sql
var TaskSchema []byte

// go:embed clean.sql
var CleanConfig []byte
