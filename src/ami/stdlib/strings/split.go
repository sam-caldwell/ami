package strings

import stdstrings "strings"

// Split slices s into all substrings separated by sep and returns a slice of the substrings between those separators.
func Split(s, sep string) []string { return stdstrings.Split(s, sep) }

