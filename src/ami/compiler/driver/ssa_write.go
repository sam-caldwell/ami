package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// writeSSADebug emits a simple SSA debug view for straight-line code by versioning definitions.
func writeSSADebug(pkg, unit string, m ir.Module) (string, error) {
    out := ssaUnit{Schema: "ssa.v1", Package: pkg, Unit: unit}
    for _, f := range m.Functions {
        sf := ssaFunc{Name: f.Name}
        versions := map[string]int{}
        for _, b := range f.Blocks {
            for _, ins := range b.Instr {
                switch v := ins.(type) {
                case ir.Var:
                    n := versions[v.Name]
                    sf.Defs = append(sf.Defs, ssaDef{Name: v.Name, Version: n, SSAName: v.Name + "#" + itoa(n), Type: v.Type})
                    versions[v.Name] = n + 1
                case ir.Assign:
                    name := v.DestID
                    n := versions[name]
                    typ := ""
                    if v.Src.Type != "" { typ = v.Src.Type }
                    sf.Defs = append(sf.Defs, ssaDef{Name: name, Version: n, SSAName: name + "#" + itoa(n), Type: typ})
                    versions[name] = n + 1
                }
            }
        }
        out.Functions = append(out.Functions, sf)
    }
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(out, "", "  ")
    if err != nil { return "", err }
    path := filepath.Join(dir, unit+".ssa.json")
    if err := os.WriteFile(path, b, 0o644); err != nil { return "", err }
    return path, nil
}

