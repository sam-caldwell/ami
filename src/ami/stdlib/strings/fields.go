package strings

import stdstrings "strings"

// Fields splits the string s around each instance of one or more consecutive white space characters.
func Fields(s string) []string { return stdstrings.Fields(s) }

