package types

import "unicode"

// splitTop splits "A,B" at top-level, respecting nested generics.
func isTypeVarName(s string) bool {
    if len(s) == 1 {
        r := rune(s[0])
        return unicode.IsUpper(r)
    }
    return false
}

