package driver

import (
    "encoding/json"
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

type irTypesIndexUnit struct {
    Unit  string   `json:"unit"`
    Types []string `json:"types"`
}

type irTypesIndex struct {
    Schema  string             `json:"schema"`
    Package string             `json:"package"`
    Units   []irTypesIndexUnit `json:"units"`
}

// collectTypes gathers a unique, sorted list of type names used in the module.
func collectTypes(m ir.Module) []string {
    seen := map[string]bool{}
    add := func(t string) { if t != "" { seen[t] = true } }
    // params/results
    for _, f := range m.Functions {
        for _, v := range f.Params { add(v.Type) }
        for _, v := range f.Results { add(v.Type) }
        for _, b := range f.Blocks {
            for _, ins := range b.Instr {
                switch x := ins.(type) {
                case ir.Var:
                    add(x.Type)
                    if x.Init != nil { add(x.Init.Type) }
                    add(x.Result.Type)
                case ir.Assign:
                    add(x.Src.Type)
                case ir.Return:
                    for _, v := range x.Values { add(v.Type) }
                case ir.Defer:
                    for _, a := range x.Expr.Args { add(a.Type) }
                    if x.Expr.Result != nil { add(x.Expr.Result.Type) }
                case ir.Expr:
                    for _, a := range x.Args { add(a.Type) }
                    if x.Result != nil { add(x.Result.Type) }
                }
            }
        }
    }
    // build sorted list
    out := make([]string, 0, len(seen))
    for k := range seen { out = append(out, k) }
    sortStrings(out)
    return out
}

// writeIRTypesIndex writes ir.types.index.json for a package.
func writeIRTypesIndex(pkg string, units []irTypesIndexUnit) (string, error) {
    idx := irTypesIndex{Schema: "ir.types.index.v1", Package: pkg, Units: units}
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(idx, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, "ir.types.index.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}

// sortStrings sorts a slice of strings in-place (local helper to avoid extra imports).
func sortStrings(a []string) {
    // simple insertion sort (small lists typical)
    for i := 1; i < len(a); i++ {
        j := i
        for j > 0 && a[j] < a[j-1] {
            a[j], a[j-1] = a[j-1], a[j]
            j--
        }
    }
}

