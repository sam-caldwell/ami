package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
)

// BuildFunctionSignatures constructs a map of function name to types.Function
// using parameter and result type strings from the AST. It ignores unnamed funcs.
func BuildFunctionSignatures(f *ast.File) map[string]types.Function {
    out := map[string]types.Function{}
    if f == nil { return out }
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            if fn.Name == "" { continue }
            out[fn.Name] = types.BuildFunction(fn)
        }
    }
    return out
}

