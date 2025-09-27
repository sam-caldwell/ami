package logschema

import (
    "bytes"
    "encoding/json"
    "sort"
    "time"
)

// Record is the stable log.v1 schema for logging events.
type Record struct {
    Timestamp time.Time
    Level     Level
    Message   string
    Package   string
    Fields    map[string]any
    Pipeline  string // optional
    Node      string // optional
}

// MarshalJSON ensures stable key ordering: timestamp, level, package, message, fields, pipeline, node
func (r Record) MarshalJSON() ([]byte, error) {
    var buf bytes.Buffer
    buf.WriteByte('{')
    // schema
    buf.WriteString("\"schema\":\"log.v1\"")
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
    // package
    if r.Package != "" {
        buf.WriteByte(',')
        buf.WriteString("\"package\":")
        pb, _ := json.Marshal(r.Package)
        buf.Write(pb)
    }
    // message
    buf.WriteByte(',')
    buf.WriteString("\"message\":")
    mb, _ := json.Marshal(r.Message)
    buf.Write(mb)
    // fields
    if len(r.Fields) > 0 {
        buf.WriteByte(',')
        buf.WriteString("\"fields\":{")
        keys := make([]string, 0, len(r.Fields))
        for k := range r.Fields { keys = append(keys, k) }
        sort.Strings(keys)
        for i, k := range keys {
            if i > 0 { buf.WriteByte(',') }
            kb, _ := json.Marshal(k)
            buf.Write(kb)
            buf.WriteByte(':')
            vb, _ := json.Marshal(r.Fields[k])
            buf.Write(vb)
        }
        buf.WriteByte('}')
    }
    // pipeline
    if r.Pipeline != "" {
        buf.WriteByte(',')
        buf.WriteString("\"pipeline\":")
        pb, _ := json.Marshal(r.Pipeline)
        buf.Write(pb)
    }
    // node
    if r.Node != "" {
        buf.WriteByte(',')
        buf.WriteString("\"node\":")
        nb, _ := json.Marshal(r.Node)
        buf.Write(nb)
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

