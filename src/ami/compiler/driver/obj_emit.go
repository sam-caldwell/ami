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
    Symbols   []objSym  `json:"symbols"`
    Relocs    []objRel  `json:"relocs"`
}

type objSym struct {
    Name string `json:"name"`
    Kind string `json:"kind"` // e.g., "func", "data"
    Addr uint64 `json:"addr"`  // placeholder address (0 for scaffold)
}

type objRel struct {
    Off   uint64 `json:"off"`
    Type  string `json:"type"`
    Sym   string `json:"sym"`
    Add   int64  `json:"add"`
}

// writeObjectStub emits a placeholder object file under build/obj/<package>/<unit>.o
// capturing function names. This serves as a stable scaffold for later real codegen.
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
