package sem

import "unicode"

func validFieldName(s string) bool {
    if s == "" { return false }
    // allow dot-separated identifiers
    start := 0
    for i, r := range s {
        if r == '.' {
            if i == start { return false }
            start = i + 1
            continue
        }
        if i == start { // start of a part
            if !(r == '_' || unicode.IsLetter(r)) { return false }
        } else {
            if !(r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)) { return false }
        }
    }
    return start < len(s)
}

