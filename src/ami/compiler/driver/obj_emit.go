package driver

import (
    "encoding/json"
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// objUnit is a tiny, deterministic object stub used for early codegen scaffolding.
type objUnit struct {
    Schema    string   `json:"schema"`
    Package   string   `json:"package"`
    Unit      string   `json:"unit"`
    Functions []string `json:"functions"`
}

// writeObjectStub emits a placeholder object file under build/obj/<package>/<unit>.o
// capturing function names. This serves as a stable scaffold for later real codegen.
func writeObjectStub(pkg, unit string, m ir.Module) (string, error) {
    base := filepath.Join("build", "obj", pkg)
    if err := os.MkdirAll(base, 0o755); err != nil { return "", err }
    out := filepath.Join(base, unit+".o")
    var fns []string
    for _, f := range m.Functions { fns = append(fns, f.Name) }
    obj := objUnit{Schema: "obj.v1", Package: pkg, Unit: unit, Functions: fns}
    b, err := json.MarshalIndent(obj, "", "  ")
    if err != nil { return "", err }
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}

