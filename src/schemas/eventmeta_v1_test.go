package schemas

import (
    "encoding/json"
    "testing"
)

func TestEventMetaV1_ValidateAndMarshal(t *testing.T) {
    m := &EventMetaV1{Schema: "eventmeta.v1", Package: "p", File: "f.ami", ImmutablePayload: true,
        Fields: []EventMetaFieldV1{{Name:"id", Type:"string"}, {Name:"timestamp", Type:"iso8601"}, {Name:"attempt", Type:"int"}},
        Trace: &TraceContextV1{Traceparent: EventMetaFieldV1{Name:"traceparent", Type:"string"}, Tracestate: EventMetaFieldV1{Name:"tracestate", Type:"string"}},
    }
    if err := m.Validate(); err != nil { t.Fatalf("validate: %v", err) }
    b, err := json.Marshal(m)
    if err != nil || len(b) == 0 { t.Fatalf("marshal: %v", err) }
    var got EventMetaV1
    if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("unmarshal: %v", err) }
    if got.Schema != "eventmeta.v1" || !got.ImmutablePayload || len(got.Fields) < 3 { t.Fatalf("round-trip mismatch: %+v", got) }
    if got.Trace == nil || got.Trace.Traceparent.Type != "string" { t.Fatalf("missing trace context: %+v", got.Trace) }
}

func TestEventMetaV1_Validate_SadSchema(t *testing.T) {
    m := &EventMetaV1{Schema: "bad"}
    if err := m.Validate(); err == nil { t.Fatalf("expected schema error") }
}
