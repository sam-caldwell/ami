package strings

import stdstrings "strings"

// Replace returns a copy of the string s with the first n non-overlapping instances of old replaced by new.
// If n < 0, there is no limit on the number of replacements.
func Replace(s, old, new string, n int) string { return stdstrings.Replace(s, old, new, n) }

