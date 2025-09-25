package bytes

import stdbytes "bytes"

// Split slices s into all subslices separated by sep and returns a slice of the subslices between those separators.
func Split(s, sep []byte) [][]byte { return stdbytes.Split(s, sep) }
