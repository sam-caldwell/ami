package io

import "os"

// Open opens an existing file for reading (like os.Open).
func Open(name string) (*FHO, error) {
    if err := guardFS(); err != nil { return nil, err }
    f, err := os.Open(name)
    if err != nil { return nil, err }
    return &FHO{f: f, name: f.Name()}, nil
}

