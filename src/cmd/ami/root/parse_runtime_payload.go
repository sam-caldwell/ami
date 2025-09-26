package root

import "strings"

// parseRuntimePayload parses key=value pairs where value may be quoted or a JSON object/array.
// Supported keys: pipeline, input, expect_output, expect_error, timeout
func parseRuntimePayload(s string) map[string]string {
    out := map[string]string{}
    i := 0
    // helper to skip spaces
    skip := func() {
        for i < len(s) && s[i] == ' ' {
            i++
        }
    }
    for i < len(s) {
        skip()
        if i >= len(s) { break }
        // parse key
        ks := i
        for i < len(s) && s[i] != '=' && s[i] != ' ' { i++ }
        if i >= len(s) || s[i] != '=' {
            // skip to next space if malformed
            for i < len(s) && s[i] != ' ' { i++ }
            continue
        }
        key := strings.TrimSpace(s[ks:i])
        i++ // skip '='
        skip()
        if i >= len(s) {
            out[key] = ""
            break
        }
        // parse value
        var val string
        switch s[i] {
        case '\'', '"':
            quote := s[i]
            i++
            vs := i
            for i < len(s) && s[i] != quote { i++ }
            if i <= len(s) { val = s[vs:i] }
            if i < len(s) { i++ }
        case '{':
            // read balanced braces
            depth := 0
            vs := i
            for i < len(s) {
                if s[i] == '{' { depth++ }
                if s[i] == '}' {
                    depth--
                    if depth == 0 { i++; break }
                }
                i++
            }
            val = s[vs:i]
        case '[':
            depth := 0
            vs := i
            for i < len(s) {
                if s[i] == '[' { depth++ }
                if s[i] == ']' {
                    depth--
                    if depth == 0 { i++; break }
                }
                i++
            }
            val = s[vs:i]
        default:
            vs := i
            for i < len(s) && s[i] != ' ' { i++ }
            val = s[vs:i]
        }
        out[strings.ToLower(key)] = strings.TrimSpace(val)
        skip()
    }
    return out
}

