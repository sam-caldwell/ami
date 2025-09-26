package os

import stdos "os"

// MkdirAll creates a directory named path, along with any necessary parents, and returns nil, or else returns an error.
func MkdirAll(path string, perm stdos.FileMode) error { return stdos.MkdirAll(path, perm) }

