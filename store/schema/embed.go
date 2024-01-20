package schema

import _ "embed"

//go:embed task.schema.sql
var TaskSchema []byte

//go:embed read.schema.sql
var ReadSchema []byte

// go:embed clean.sql
var CleanConfig []byte
