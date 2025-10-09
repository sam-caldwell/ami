package llvm

// encodeCString encodes s into a C-style bytes string suitable for LLVM c"..." with \00 terminator.
func encodeCString(s string) string {
    // append terminator
    b := make([]byte, 0, len(s)+1)
    for i := 0; i < len(s); i++ {
        c := s[i]
        if c == '\\' || c == '"' { b = append(b, '\\', c); continue }
        // limit to ASCII printing; otherwise use \xx hex escapes
        if c < 0x20 || c > 0x7e {
            // hex escape
            hi := "0123456789ABCDEF"[c>>4]
            lo := "0123456789ABCDEF"[c&0xF]
            b = append(b, '\\', 'x', hi, lo)
            continue
        }
        b = append(b, c)
    }
    b = append(b, '\\', '0', '0')
    return string(b)
}
