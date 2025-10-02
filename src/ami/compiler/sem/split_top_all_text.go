package sem

import "strings"

// splitTopAllText splits on commas at top level, respecting nested '<>', '{}', '()' and quotes.
func splitTopAllText(s string) []string {
    var parts []string
    depthAngle, depthBrace, depthParen := 0, 0, 0
    inQuote := byte(0)
    last := 0
    for i := 0; i < len(s); i++ {
        c := s[i]
        if inQuote != 0 {
            if c == inQuote { inQuote = 0 }
            continue
        }
        switch c {
        case '\'', '"':
            inQuote = c
        case '<':
            depthAngle++
        case '>':
            if depthAngle > 0 { depthAngle-- }
        case '{':
            depthBrace++
        case '}':
            if depthBrace > 0 { depthBrace-- }
        case '(':
            depthParen++
        case ')':
            if depthParen > 0 { depthParen-- }
        case ',':
            if depthAngle == 0 && depthBrace == 0 && depthParen == 0 {
                parts = append(parts, strings.TrimSpace(s[last:i]))
                last = i + 1
            }
        }
    }
    tail := strings.TrimSpace(s[last:])
    if tail != "" { parts = append(parts, tail) }
    return parts
}

