package io

import "os"

// OpenFile is a general opener (like os.OpenFile) returning an FHO.
func OpenFile(name string, flag int, perm os.FileMode) (*FHO, error) {
    if err := guardFS(); err != nil { return nil, err }
    f, err := os.OpenFile(name, flag, perm)
    if err != nil { return nil, err }
    return &FHO{f: f, name: f.Name()}, nil
}

