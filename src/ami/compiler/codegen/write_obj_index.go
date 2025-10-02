package codegen

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// WriteObjIndex writes the index JSON to build/obj/<package>/index.json.
func WriteObjIndex(idx ObjIndex) error {
    base := filepath.Join("build", "obj", idx.Package)
    if err := os.MkdirAll(base, 0o755); err != nil { return err }
    b, err := json.MarshalIndent(idx, "", "  ")
    if err != nil { return err }
    return os.WriteFile(filepath.Join(base, "index.json"), b, 0o644)
}

