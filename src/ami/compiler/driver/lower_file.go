package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerFile lowers functions found in an AST file into a single IR module.
func lowerFile(pkg string, f *ast.File, params map[string][]string, results map[string][]string, paramNames map[string][]string) ir.Module {
    // signature maps are provided by caller (compile phase)
    var fns []ir.Function
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            fns = append(fns, lowerFuncDecl(fn, results, params, paramNames))
        }
    }
    return ir.Module{Package: pkg, Functions: fns}
}
