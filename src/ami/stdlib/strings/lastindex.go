package strings

import stdstrings "strings"

// LastIndex returns the index of the last instance of substr in s, or -1 if substr is not present in s.
func LastIndex(s, substr string) int { return stdstrings.LastIndex(s, substr) }
