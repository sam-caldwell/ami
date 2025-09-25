package filepath

import stdpath "path"

// Ext returns the file name extension used by path.
func Ext(path string) string { return stdpath.Ext(normalizeSeparators(path)) }

