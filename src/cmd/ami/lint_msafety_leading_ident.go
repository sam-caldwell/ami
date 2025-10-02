package main

// leadingIdent returns the leading identifier from s (letters/digits/underscore allowed in scan).
func leadingIdent(s string) string {
    i := 0
    for i < len(s) {
        c := s[i]
        if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
            i++
            continue
        }
        break
    }
    return s[:i]
}

