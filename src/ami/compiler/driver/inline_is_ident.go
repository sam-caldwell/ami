package driver

// isIdent reports whether s is a simple identifier (letters, digits, underscore), starting with a letter or underscore.
func isIdent(s string) bool {
    if s == "" { return false }
    for i := 0; i < len(s); i++ {
        ch := s[i]
        if i == 0 {
            if !(ch == '_' || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')) { return false }
        } else {
            if !((ch == '_') || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) { return false }
        }
    }
    return true
}

