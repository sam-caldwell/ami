package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerFile lowers functions found in an AST file into a single IR module.
func lowerFile(pkg string, f *ast.File) ir.Module {
    var fns []ir.Function
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            fns = append(fns, lowerFuncDecl(fn))
        }
    }
    return ir.Module{Package: pkg, Functions: fns}
}

