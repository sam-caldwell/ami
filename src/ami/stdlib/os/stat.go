package os

import stdos "os"

// Stat returns a FileInfo describing the named file.
func Stat(name string) (stdos.FileInfo, error) { return stdos.Stat(name) }

