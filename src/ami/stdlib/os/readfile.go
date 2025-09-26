package os

import stdos "os"

// ReadFile reads the named file and returns the contents.
func ReadFile(name string) ([]byte, error) { return stdos.ReadFile(name) }

