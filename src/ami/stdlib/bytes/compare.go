package bytes

import stdbytes "bytes"

// Compare returns an integer comparing two byte slices lexicographically.
func Compare(a, b []byte) int { return stdbytes.Compare(a, b) }
