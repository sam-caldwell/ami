package strings

import stdstrings "strings"

// EqualFold reports whether s and t, interpreted as UTF-8 strings, are equal under Unicode case-folding.
func EqualFold(s, t string) bool { return stdstrings.EqualFold(s, t) }
