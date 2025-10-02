package logger

import (
    logschema "github.com/sam-caldwell/ami/src/schemas/log"
)

// redactLogV1LineAdvanced applies allow/deny + redact behavior to a log.v1 JSON line.
// Order: allowlist (keep only if provided), denylist (drop), redact exact keys, redact by prefix.
func redactLogV1LineAdvanced(line []byte, allow, deny, redact, prefixes []string) ([]byte, bool) {
    var rec logschema.Record
    if err := jsonUnmarshal(line, &rec); err != nil { return line, false }
    if len(rec.Fields) == 0 { return line, false }
    // allow
    out := map[string]any{}
    if len(allow) > 0 {
        for _, k := range allow { if v, ok := rec.Fields[k]; ok { out[k] = v } }
    } else {
        for k, v := range rec.Fields { out[k] = v }
    }
    // deny
    for _, k := range deny { delete(out, k) }
    // redact exact
    for _, k := range redact { if _, ok := out[k]; ok { out[k] = "[REDACTED]" } }
    // redact prefixes
    if len(prefixes) > 0 {
        for k := range out {
            for _, p := range prefixes {
                if len(p) > 0 && len(k) >= len(p) && k[:len(p)] == p { out[k] = "[REDACTED]"; break }
            }
        }
    }
    rec.Fields = out
    b, err := rec.MarshalJSON(); if err != nil { return line, false }
    return append(b, '\n'), true
}

