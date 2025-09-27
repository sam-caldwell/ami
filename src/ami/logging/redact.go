package logging

// filterRedactFields applies allow/deny filtering and redactions to a copy of fields.
// Order of operations: allowlist (keep), then denylist (drop), then redact exact keys, then redact by prefix.
func filterRedactFields(fields map[string]any, allow, deny, redact, prefixes []string) map[string]any {
    if len(fields) == 0 { return fields }
    out := make(map[string]any, len(fields))

    // allowlist: if present, keep only allowed keys
    if len(allow) > 0 {
        for _, k := range allow {
            if v, ok := fields[k]; ok { out[k] = v }
        }
    } else {
        for k, v := range fields { out[k] = v }
    }

    // denylist: drop these keys
    if len(deny) > 0 {
        for _, k := range deny { delete(out, k) }
    }

    // redact exact keys
    for _, k := range redact {
        if _, ok := out[k]; ok { out[k] = "[REDACTED]" }
    }

    // redact by prefix
    if len(prefixes) > 0 {
        for k := range out {
            for _, p := range prefixes {
                if len(p) > 0 && len(k) >= len(p) && k[:len(p)] == p {
                    out[k] = "[REDACTED]"
                    break
                }
            }
        }
    }
    return out
}
