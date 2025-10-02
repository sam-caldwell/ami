package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// writeIRSymbolsIndex writes ir.symbols.index.json for a package.
func writeIRSymbolsIndex(pkg string, units []irSymbolsIndexUnit) (string, error) {
    idx := irSymbolsIndex{Schema: "ir.symbols.index.v1", Package: pkg, Units: units}
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(idx, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, "ir.symbols.index.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}

