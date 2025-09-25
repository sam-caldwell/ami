package bytes

import stdbytes "bytes"

// Contains reports whether subslice is within b.
func Contains(b, subslice []byte) bool { return stdbytes.Contains(b, subslice) }

