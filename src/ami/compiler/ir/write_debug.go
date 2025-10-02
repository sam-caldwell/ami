package ir

import (
    "os"
    "path/filepath"
)

// WriteDebug writes module JSON under build/debug/ir/<pkg>.json
func WriteDebug(m Module) error {
    path := filepath.Join("build", "debug", "ir")
    if err := os.MkdirAll(path, 0o755); err != nil { return err }
    b, err := EncodeModule(m)
    if err != nil { return err }
    fname := filepath.Join(path, m.Package+".json")
    return os.WriteFile(fname, b, 0o644)
}

