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
    // collect directives from pragmas
    var dirs []ir.Directive
    if f != nil {
        for _, pr := range f.Pragmas {
            dirs = append(dirs, ir.Directive{Domain: pr.Domain, Key: pr.Key, Value: pr.Value, Args: append([]string(nil), pr.Args...), Params: pr.Params})
        }
    }
    return ir.Module{Package: pkg, Functions: fns, Directives: dirs}
}
