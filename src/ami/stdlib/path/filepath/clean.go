package filepath

import stdpath "path"

// Clean returns the shortest path name equivalent to path
// by purely lexical processing using '/' separators.
func Clean(path string) string { return stdpath.Clean(normalizeSeparators(path)) }
