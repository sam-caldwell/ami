package logging

import (
    "bytes"
    "encoding/json"
    "sort"
    "strings"
    "time"
)

// Record represents a structured log event.
type Record struct {
    Timestamp time.Time
    Level     Level
    Message   string
    Package   string
    Fields    map[string]any
}

// normalizeMsg converts CRLF to LF to keep outputs deterministic.
func normalizeMsg(s string) string {
    return strings.ReplaceAll(s, "\r\n", "\n")
}

// MarshalJSON ensures deterministic key ordering and stable output.
func (r Record) MarshalJSON() ([]byte, error) {
    var buf bytes.Buffer
    buf.WriteByte('{')
    // Order: timestamp, level, package, message, fields
    // timestamp
    buf.WriteString("\"timestamp\":")
    ts := iso8601UTCms(r.Timestamp)
    tsb, _ := json.Marshal(ts)
    buf.Write(tsb)
    // level
    buf.WriteByte(',')
    buf.WriteString("\"level\":")
    lvb, _ := json.Marshal(string(r.Level))
    buf.Write(lvb)
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
    mb, _ := json.Marshal(normalizeMsg(r.Message))
    buf.Write(mb)
    // fields
    if len(r.Fields) > 0 {
        buf.WriteByte(',')
        buf.WriteString("\"fields\":{")
        keys := make([]string, 0, len(r.Fields))
        for k := range r.Fields {
            keys = append(keys, k)
        }
        sort.Strings(keys)
        for i, k := range keys {
            if i > 0 {
                buf.WriteByte(',')
            }
            kb, _ := json.Marshal(k)
            buf.Write(kb)
            buf.WriteByte(':')
            vb, _ := json.Marshal(r.Fields[k])
            buf.Write(vb)
        }
        buf.WriteByte('}')
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

