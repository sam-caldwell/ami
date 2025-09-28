package driver

import (
    "encoding/json"
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// writeExportsDebug writes a simple exports table for a unit listing IR function names.
// Path: build/debug/link/<pkg>/<unit>.exports.json
func writeExportsDebug(pkg, unit string, m ir.Module) (string, error) {
    dir := filepath.Join("build", "debug", "link", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    path := filepath.Join(dir, unit+".exports.json")
    // Collect exported names (public functions only: Name starts with uppercase)
    exports := make([]string, 0, len(m.Functions))
    for _, f := range m.Functions {
        if len(f.Name) > 0 && f.Name[0] >= 'A' && f.Name[0] <= 'Z' {
            exports = append(exports, f.Name)
        }
    }
    obj := map[string]any{"schema": "exports.v1", "package": pkg, "unit": unit, "functions": exports}
    b, _ := json.Marshal(obj)
    return path, os.WriteFile(path, b, 0o644)
}

