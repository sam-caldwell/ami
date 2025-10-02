package llvm

import (
    "os"
    "path/filepath"
)

// WriteRuntimeLL writes the runtime LLVM IR text to the given directory and returns the file path.
func WriteRuntimeLL(dir, triple string, withMain bool) (string, error) {
    if triple == "" { triple = DefaultTriple }
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    path := filepath.Join(dir, "runtime.ll")
    return path, os.WriteFile(path, []byte(RuntimeLL(triple, withMain)), 0o644)
}

