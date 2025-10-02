package main

import "strings"

// parseKV parses a simple space-delimited list of key=value tokens. Values may be wrapped in single or double quotes.
func parseKV(s string) map[string]string {
    out := map[string]string{}
    parts := strings.Fields(s)
    for _, p := range parts {
        if i := strings.IndexByte(p, '='); i > 0 {
            k := p[:i]
            v := p[i+1:]
            v = strings.Trim(v, `"'`)
            out[k] = v
        }
    }
    return out
}

