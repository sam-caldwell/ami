package io

import "os"

// Create creates or truncates a file for writing (like os.Create).
func Create(name string) (*FHO, error) {
    if err := guardFS(); err != nil { return nil, err }
    f, err := os.Create(name)
    if err != nil { return nil, err }
    return &FHO{f: f, name: f.Name()}, nil
}

