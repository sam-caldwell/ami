package os

import stdos "os"

// Mkdir creates a new directory with the specified name and permission bits.
func Mkdir(name string, perm stdos.FileMode) error { return stdos.Mkdir(name, perm) }

