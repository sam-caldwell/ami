package diag

import (
    "bytes"
    "encoding/json"
    "sort"
    "time"
)

// Record is the stable diag.v1 object with deterministic JSON ordering.
type Record struct {
    Timestamp time.Time
    Level     Level
    Code      string
    Message   string
    Package   string
    File      string
    Pos       *Position
    EndPos    *Position
    Data      map[string]any
}

// MarshalJSON renders keys in a stable order, matching diag.v1 schema.
func (r Record) MarshalJSON() ([]byte, error) {
    var buf bytes.Buffer
    buf.WriteByte('{')
    // schema constant first
    buf.WriteString("\"schema\":\"diag.v1\"")
    // timestamp
    buf.WriteByte(',')
    buf.WriteString("\"timestamp\":")
    ts, _ := json.Marshal(r.Timestamp.UTC().Format("2006-01-02T15:04:05.000Z"))
    buf.Write(ts)
    // level
    buf.WriteByte(',')
    buf.WriteString("\"level\":")
    lv, _ := json.Marshal(string(r.Level))
    buf.Write(lv)
    // code
    buf.WriteByte(',')
    buf.WriteString("\"code\":")
    cb, _ := json.Marshal(r.Code)
    buf.Write(cb)
    // message
    buf.WriteByte(',')
    buf.WriteString("\"message\":")
    mb, _ := json.Marshal(r.Message)
    buf.Write(mb)
    // package (optional)
    if r.Package != "" {
        buf.WriteByte(',')
        buf.WriteString("\"package\":")
        pb, _ := json.Marshal(r.Package)
        buf.Write(pb)
    }
    // file (optional)
    if r.File != "" {
        buf.WriteByte(',')
        buf.WriteString("\"file\":")
        fb, _ := json.Marshal(r.File)
        buf.Write(fb)
    }
    // pos (optional)
    if r.Pos != nil {
        buf.WriteByte(',')
        buf.WriteString("\"pos\":")
        p, _ := json.Marshal(r.Pos)
        buf.Write(p)
    }
    // endPos (optional)
    if r.EndPos != nil {
        buf.WriteByte(',')
        buf.WriteString("\"endPos\":")
        ep, _ := json.Marshal(r.EndPos)
        buf.Write(ep)
    }
    // data (optional) deterministic map ordering
    if len(r.Data) > 0 {
        buf.WriteByte(',')
        buf.WriteString("\"data\":{")
        keys := make([]string, 0, len(r.Data))
        for k := range r.Data { keys = append(keys, k) }
        sort.Strings(keys)
        for i, k := range keys {
            if i > 0 { buf.WriteByte(',') }
            kb, _ := json.Marshal(k)
            buf.Write(kb)
            buf.WriteByte(':')
            vb, _ := json.Marshal(r.Data[k])
            buf.Write(vb)
        }
        buf.WriteByte('}')
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

