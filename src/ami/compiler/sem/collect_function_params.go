package sem

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// collectFunctionParams builds a map of function name to declared parameter types (textual), for transfer checks.
func collectFunctionParams(f *ast.File) map[string][]string {
    out := map[string][]string{}
    if f == nil { return out }
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            var ps []string
            for _, p := range fn.Params { ps = append(ps, p.Type) }
            out[fn.Name] = ps
        }
    }
    return out
}

