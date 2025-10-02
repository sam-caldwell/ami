package sem

import "strings"

func parseStructFieldsText(s string) (map[string]string, bool) {
    s = strings.TrimSpace(s)
    if !isStructText(s) { return nil, false }
    body := s[len("Struct{") : len(s)-1]
    out := map[string]string{}
    if strings.TrimSpace(body) == "" { return out, true }
    parts := splitTopAllText(body)
    for _, p := range parts {
        if i := strings.IndexByte(p, ':'); i > 0 {
            name := strings.TrimSpace(p[:i])
            ty := strings.TrimSpace(p[i+1:])
            if name != "" { out[name] = ty }
        } else {
            return nil, false
        }
    }
    return out, true
}

