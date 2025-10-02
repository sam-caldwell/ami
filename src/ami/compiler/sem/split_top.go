package sem

import "strings"

func splitTop(s string) []string {
    var parts []string
    depth := 0
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
            depth++
        case '>':
            if depth > 0 { depth-- }
        case ',':
            if depth == 0 {
                parts = append(parts, strings.TrimSpace(s[last:i]))
                last = i + 1
            }
        }
    }
    tail := strings.TrimSpace(s[last:])
    if tail != "" { parts = append(parts, tail) }
    return parts
}

