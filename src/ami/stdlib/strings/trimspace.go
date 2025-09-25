package strings

import stdstrings "strings"

// TrimSpace returns a slice of the string s, with all leading and trailing white space removed, as defined by Unicode.
func TrimSpace(s string) string { return stdstrings.TrimSpace(s) }
