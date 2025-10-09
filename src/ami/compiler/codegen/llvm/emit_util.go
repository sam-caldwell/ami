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
// (moved: see emit_util_itoa.go and emit_util_meta.go)
