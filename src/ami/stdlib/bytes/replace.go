package bytes

import stdbytes "bytes"

// Replace returns a copy of s with the first n non-overlapping instances of old replaced by new.
// If n < 0, there is no limit on the number of replacements.
func Replace(s, old, new []byte, n int) []byte { return stdbytes.Replace(s, old, new, n) }
