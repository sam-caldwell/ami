package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// collectFunctionResults builds a map of function name to declared result types.
func collectFunctionResults(f *ast.File) map[string][]string {
    out := map[string][]string{}
    if f == nil { return out }
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            if len(fn.Results) == 0 { continue }
            rs := make([]string, len(fn.Results))
            for i, r := range fn.Results { rs[i] = r.Type }
            out[fn.Name] = rs
        }
    }
    return out
}

