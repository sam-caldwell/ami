package ir

// EventMeta describes the standardized event metadata fields captured by the
// runtime and surfaced for observability. This does not change codegen but is
// included in IR debug artifacts to make expectations explicit and testable.
type EventMeta struct {
    Schema string   `json:"schema"`             // e.g., "eventmeta.v1"
    Fields []string `json:"fields,omitempty"`  // canonical field names
}

