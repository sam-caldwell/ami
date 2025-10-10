package llvm

// encodeCString encodes s into an LLVM IR-compatible c"..." byte string, ending with \00.
// It uses octal escapes (\NNN) and avoids C-style \" for quotes to prevent IR parse issues.
func encodeCString(s string) string {
    // append terminator
    b := make([]byte, 0, len(s)+4)
    for i := 0; i < len(s); i++ {
        c := s[i]
        switch c {
        case '"':
            // quote as octal \22
            b = append(b, '\\', '2', '2')
            continue
        case '\\':
            // backslash as octal \5C
            b = append(b, '\\', '5', 'C')
            continue
        }
        // printable ASCII
        if c >= 0x20 && c <= 0x7e {
            b = append(b, c)
            continue
        }
        // other bytes: 3-digit octal escape
        o2 := '0' + ((c >> 6) & 0x7)
        o1 := '0' + ((c >> 3) & 0x7)
        o0 := '0' + (c & 0x7)
        b = append(b, '\\', byte(o2), byte(o1), byte(o0))
    }
    // trailing NUL
    b = append(b, '\\', '0', '0')
    return string(b)
}
