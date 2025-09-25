package bytes

import stdbytes "bytes"

// Index returns the index of the first instance of sep in b, or -1 if sep is not present in b.
func Index(b, sep []byte) int { return stdbytes.Index(b, sep) }
