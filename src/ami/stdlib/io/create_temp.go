package io

import (
    "errors"
    "os"
    "path/filepath"
    "strings"
)

// CreateTemp creates a new temporary file under the system temp dir.
// Optional args: [dir] or [dir, suffix]. The dir is relative to the system temp directory.
func CreateTemp(args ...string) (*FHO, error) {
    if err := guardFS(); err != nil { return nil, err }
    base := os.TempDir()
    var rel, suffix string
    switch len(args) {
    case 0:
    case 1:
        rel = args[0]
    case 2:
        rel = args[0]
        suffix = args[1]
    default:
        return nil, errors.New("CreateTemp: invalid arguments; want none, [dir], or [dir,suffix]")
    }
    dir := base
    if strings.TrimSpace(rel) != "" { dir = filepath.Join(base, rel) }
    if err := os.MkdirAll(dir, 0o755); err != nil { return nil, err }
    pattern := "ami-*" + suffix
    f, err := os.CreateTemp(dir, pattern)
    if err != nil { return nil, err }
    return &FHO{f: f, name: f.Name()}, nil
}

