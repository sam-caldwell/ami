package bytes

import stdbytes "bytes"

// LastIndex returns the index of the last instance of sep in b, or -1 if sep is not present in b.
func LastIndex(b, sep []byte) int { return stdbytes.LastIndex(b, sep) }
