package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/ir"

// collectTypes gathers a unique, sorted list of type names used in the module.
func collectTypes(m ir.Module) []string {
    seen := map[string]bool{}
    add := func(t string) { if t != "" { seen[t] = true } }
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
    out := make([]string, 0, len(seen))
    for k := range seen { out = append(out, k) }
    sortStrings(out)
    return out
}

