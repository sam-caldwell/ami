package logger

import (
    logschema "github.com/sam-caldwell/ami/src/schemas/log"
)

// redactLogV1Line attempts to parse a log.v1 JSON line and apply redactions to fields.
// Returns the possibly modified line and true if redaction was applied via re-encoding.
func redactLogV1Line(line []byte, keys, prefixes []string) ([]byte, bool) {
    var rec logschema.Record
    if err := jsonUnmarshal(line, &rec); err != nil {
        return line, false
    }
    if len(rec.Fields) == 0 { return line, false }
    // exact keys
    for _, k := range keys {
        if _, ok := rec.Fields[k]; ok { rec.Fields[k] = "[REDACTED]" }
    }
    // prefixes
    if len(prefixes) > 0 {
        for k := range rec.Fields {
            for _, p := range prefixes {
                if len(p) > 0 && len(k) >= len(p) && k[:len(p)] == p {
                    rec.Fields[k] = "[REDACTED]"
                    break
                }
            }
        }
    }
    b, err := rec.MarshalJSON()
    if err != nil { return line, false }
    b = append(b, '\n')
    return b, true
}

// small indirection to allow testing without import cycles
var jsonUnmarshal = func(b []byte, v any) error { return stdJSONUnmarshal(b, v) }
