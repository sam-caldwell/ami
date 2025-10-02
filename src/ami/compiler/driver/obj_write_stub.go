package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// writeObjectStub emits a placeholder object file under build/obj/<package>/<unit>.o
func writeObjectStub(pkg, unit string, m ir.Module) (string, error) {
    base := filepath.Join("build", "obj", pkg)
    if err := os.MkdirAll(base, 0o755); err != nil { return "", err }
    out := filepath.Join(base, unit+".o")
    var fns []string
    var syms []objSym
    for _, f := range m.Functions {
        fns = append(fns, f.Name)
        syms = append(syms, objSym{Name: f.Name, Kind: "func", Addr: 0})
    }
    obj := objUnit{Schema: "obj.v1", Package: pkg, Unit: unit, Functions: fns, Symbols: syms, Relocs: []objRel{}}
    b, err := json.MarshalIndent(obj, "", "  ")
    if err != nil { return "", err }
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}

