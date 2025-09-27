package events

import (
    "bytes"
    "encoding/json"
    "time"
)

// Event is the stable events.v1 object used to represent a single pipeline event.
// Fields are intentionally minimal and will expand as the spec solidifies.
type Event struct {
    ID        string
    Timestamp time.Time
    Attempt   int
    Trace     map[string]any
    Payload   any
}

// MarshalJSON renders keys in a stable order with a schema discriminator.
func (e Event) MarshalJSON() ([]byte, error) {
    var buf bytes.Buffer
    buf.WriteByte('{')
    // schema first
    buf.WriteString("\"schema\":\"events.v1\"")
    // id
    if e.ID != "" {
        buf.WriteByte(',')
        buf.WriteString("\"id\":")
        ib, _ := json.Marshal(e.ID)
        buf.Write(ib)
    }
    // timestamp
    if !e.Timestamp.IsZero() {
        buf.WriteByte(',')
        buf.WriteString("\"timestamp\":")
        tb, _ := json.Marshal(e.Timestamp.UTC().Format("2006-01-02T15:04:05.000Z"))
        buf.Write(tb)
    }
    // attempt
    if e.Attempt != 0 {
        buf.WriteByte(',')
        buf.WriteString("\"attempt\":")
        ab, _ := json.Marshal(e.Attempt)
        buf.Write(ab)
    }
    // trace (optional)
    if len(e.Trace) > 0 {
        buf.WriteByte(',')
        buf.WriteString("\"trace\":")
        tb, _ := json.Marshal(e.Trace)
        buf.Write(tb)
    }
    // payload (optional)
    if e.Payload != nil {
        buf.WriteByte(',')
        buf.WriteString("\"payload\":")
        pb, _ := json.Marshal(e.Payload)
        buf.Write(pb)
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

