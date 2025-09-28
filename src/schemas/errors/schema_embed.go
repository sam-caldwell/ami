package errors

import _ "embed"

// SchemaJSON is the embedded JSON Schema for errors.v1.
//go:embed schema.json
var SchemaJSON string

