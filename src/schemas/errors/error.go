package errors

import (
    "bytes"
    "encoding/json"
    "time"
)

// Error is the stable errors.v1 object used to represent a single
// compiler/runtime error in a compact, schema-tagged form.
// Fields align with diag.v1 where applicable.
type Error struct {
    Timestamp time.Time
    Level     string
    Code      string
    Message   string
    File      string
    Pos       *Position
    Data      map[string]any
}

// MarshalJSON renders keys in a stable order and includes a schema discriminator.
func (e Error) MarshalJSON() ([]byte, error) {
    var buf bytes.Buffer
    buf.WriteByte('{')
    // schema first
    buf.WriteString("\"schema\":\"errors.v1\"")
    // timestamp (optional)
    if !e.Timestamp.IsZero() {
        buf.WriteByte(',')
        buf.WriteString("\"timestamp\":")
        // ISO-8601 UTC with millisecond precision
        t := e.Timestamp.UTC().Truncate(time.Millisecond)
        tb, _ := json.Marshal(t.Format("2006-01-02T15:04:05.000Z"))
        buf.Write(tb)
    }
    // level
    if e.Level != "" {
        buf.WriteByte(',')
        buf.WriteString("\"level\":")
        lb, _ := json.Marshal(e.Level)
        buf.Write(lb)
    }
    // code
    if e.Code != "" {
        buf.WriteByte(',')
        buf.WriteString("\"code\":")
        cb, _ := json.Marshal(e.Code)
        buf.Write(cb)
    }
    // message
    if e.Message != "" {
        buf.WriteByte(',')
        buf.WriteString("\"message\":")
        mb, _ := json.Marshal(e.Message)
        buf.Write(mb)
    }
    // file
    if e.File != "" {
        buf.WriteByte(',')
        buf.WriteString("\"file\":")
        fb, _ := json.Marshal(e.File)
        buf.Write(fb)
    }
    // pos
    if e.Pos != nil {
        buf.WriteByte(',')
        buf.WriteString("\"pos\":")
        pb, _ := json.Marshal(e.Pos)
        buf.Write(pb)
    }
    // data
    if len(e.Data) > 0 {
        buf.WriteByte(',')
        buf.WriteString("\"data\":")
        db, _ := json.Marshal(e.Data)
        buf.Write(db)
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}
