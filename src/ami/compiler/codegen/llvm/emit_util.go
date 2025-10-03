package llvm

// sanitizeIdent converts a string into a safe suffix for LLVM global names by
// replacing characters that are not valid in identifiers with underscores.
func sanitizeIdent(s string) string {
    if s == "" { return "_" }
    b := make([]byte, len(s))
    for i := 0; i < len(s); i++ {
        c := s[i]
        if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
            b[i] = c
        } else {
            b[i] = '_'
        }
    }
    return string(b)
}

// small helper to avoid importing strconv repeatedly in minimal emitter surface
func itoa(n int) string {
    if n == 0 { return "0" }
    var buf [20]byte
    i := len(buf)
    for n > 0 {
        i--
        buf[i] = byte('0' + (n % 10))
        n /= 10
    }
    return string(buf[i:])
}

