package events

import _ "embed"

// SchemaJSON is the embedded JSON Schema for events.v1.
//go:embed schema.json
var SchemaJSON string

