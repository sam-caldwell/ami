package strings

import stdstrings "strings"

// Trim returns a slice of the string s with all leading and trailing Unicode code points contained in cutset removed.
func Trim(s, cutset string) string { return stdstrings.Trim(s, cutset) }
