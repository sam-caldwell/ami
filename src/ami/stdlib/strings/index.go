package strings

import stdstrings "strings"

// Index returns the index of the first instance of substr in s, or -1 if substr is not present in s.
func Index(s, substr string) int { return stdstrings.Index(s, substr) }
