package schemas

import "errors"

// EventMetaV1 describes the standard event lifecycle metadata available to
// all pipelines and workers. It is a debug/spec artifact emitted per unit
// to make the contract visible to tooling.
type EventMetaV1 struct {
    Schema           string             `json:"schema"`
    Timestamp        string             `json:"timestamp"`
    Package          string             `json:"package"`
    File             string             `json:"file"`
    ImmutablePayload bool               `json:"immutablePayload"`
    Fields           []EventMetaFieldV1 `json:"fields"`
    Trace            *TraceContextV1    `json:"trace,omitempty"`
}

type EventMetaFieldV1 struct {
    Name string `json:"name"`
    Type string `json:"type"` // string|iso8601|string(map)|int
    Note string `json:"note,omitempty"`
}

func (e *EventMetaV1) Validate() error {
    if e == nil { return errors.New("nil eventmeta") }
    if e.Schema == "" { e.Schema = "eventmeta.v1" }
    if e.Schema != "eventmeta.v1" { return errors.New("invalid schema") }
    if len(e.Fields) == 0 { return errors.New("no fields") }
    return nil
}

// TraceContextV1 describes the structured trace context fields following
// W3C Trace Context (traceparent/tracestate) as string-typed metadata.
type TraceContextV1 struct {
    Traceparent EventMetaFieldV1 `json:"traceparent"`
    Tracestate  EventMetaFieldV1 `json:"tracestate,omitempty"`
}
