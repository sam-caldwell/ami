package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
)

func writeBuildManifest(m BuildManifest) (string, error) {
    if m.Schema == "" { m.Schema = "manifest.v1" }
    dir := filepath.Join("build", "debug")
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(m, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, "manifest.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}

