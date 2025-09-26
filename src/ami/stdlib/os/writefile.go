package os

import stdos "os"

// WriteFile writes data to the named file with explicit permissions perm.
// If the file does not exist, WriteFile creates it with permissions perm; otherwise WriteFile truncates it before writing.
func WriteFile(name string, data []byte, perm stdos.FileMode) error {
    return stdos.WriteFile(name, data, perm)
}

