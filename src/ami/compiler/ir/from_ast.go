package ir

import (
    "sort"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// FromASTFile builds a simple IR module enumerating function declarations.
func FromASTFile(pkg, version, unit string, f *astpkg.File) Module {
    m := Module{Package: pkg, Version: version, Unit: unit, AST: f}
    for _, d := range f.Decls {
        if fd, ok := d.(astpkg.FuncDecl); ok {
            var tps []string
            for _, tp := range fd.TypeParams { tps = append(tps, tp.Name) }
            m.Functions = append(m.Functions, Function{Name: fd.Name, TypeParams: tps})
        }
    }
    sort.Slice(m.Functions, func(i, j int) bool { return m.Functions[i].Name < m.Functions[j].Name })
    return m
}
