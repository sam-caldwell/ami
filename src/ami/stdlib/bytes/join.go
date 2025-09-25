package bytes

import stdbytes "bytes"

// Join concatenates the elements of s to create a new byte slice. The separator sep is placed between elements.
func Join(s [][]byte, sep []byte) []byte { return stdbytes.Join(s, sep) }

