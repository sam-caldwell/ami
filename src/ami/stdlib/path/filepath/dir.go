package filepath

import stdpath "path"

// Dir returns all but the last element of path, typically the path's directory, using '/' semantics.
func Dir(path string) string { return stdpath.Dir(normalizeSeparators(path)) }
