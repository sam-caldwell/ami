package os

import stdos "os"

// ReadFile reads the named file and returns the contents.
func ReadFile(name string) ([]byte, error) { return stdos.ReadFile(name) }

// WriteFile writes data to the named file with explicit permissions perm.
// If the file does not exist, WriteFile creates it with permissions perm; otherwise WriteFile truncates it before writing.
func WriteFile(name string, data []byte, perm stdos.FileMode) error {
	return stdos.WriteFile(name, data, perm)
}

// Mkdir creates a new directory with the specified name and permission bits.
func Mkdir(name string, perm stdos.FileMode) error { return stdos.Mkdir(name, perm) }

// MkdirAll creates a directory named path, along with any necessary parents, and returns nil, or else returns an error.
func MkdirAll(path string, perm stdos.FileMode) error { return stdos.MkdirAll(path, perm) }

// Stat returns a FileInfo describing the named file.
func Stat(name string) (stdos.FileInfo, error) { return stdos.Stat(name) }
