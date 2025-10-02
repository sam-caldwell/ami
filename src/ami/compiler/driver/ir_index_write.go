package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// writeIRIndex writes a package-level IR index listing functions per unit.
func writeIRIndex(pkg string, units []irIndexUnit) (string, error) {
    idx := irIndex{Schema: "ir.index.v1", Package: pkg, Units: units}
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(idx, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, "ir.index.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}

