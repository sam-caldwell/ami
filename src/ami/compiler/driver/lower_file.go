package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerFile lowers functions found in an AST file into a single IR module.
func lowerFile(pkg string, f *ast.File) ir.Module {
    // collect function result types for simple call typing
    results := map[string][]string{}
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            var rs []string
            for _, r := range fn.Results { rs = append(rs, r.Type) }
            results[fn.Name] = rs
        }
    }
    var fns []ir.Function
    for _, d := range f.Decls {
        if fn, ok := d.(*ast.FuncDecl); ok {
            fns = append(fns, lowerFuncDecl(fn, results))
        }
    }
    return ir.Module{Package: pkg, Functions: fns}
}
