package filepath

import stdpath "path"

// Base returns the last element of path using '/' semantics.
func Base(path string) string { return stdpath.Base(normalizeSeparators(path)) }
