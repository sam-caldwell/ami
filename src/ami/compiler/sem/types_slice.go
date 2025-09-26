package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "strings"
)

// analyzeSliceTypes validates generic slice forms `slice<T>` for correct arity.
// Bracket slices `[]T` are represented by TypeRef{Slice:true, Name:T} and do not
// require additional constraints beyond nested type validation (e.g., maps).
func analyzeSliceTypes(f *astpkg.File) []diag.Diagnostic {
    var diags []diag.Diagnostic
    var walk func(t astpkg.TypeRef)
    walk = func(t astpkg.TypeRef) {
        if strings.ToLower(t.Name) == "slice" {
            if len(t.Args) != 1 {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_SLICE_ARITY", Message: "slice must have exactly one type argument: slice<T>"})
            }
        }
        for _, a := range t.Args {
            walk(a)
        }
    }
    for _, d := range f.Decls {
        if sd, ok := d.(astpkg.StructDecl); ok {
            for _, fld := range sd.Fields {
                walk(fld.Type)
            }
        }
        if fd, ok := d.(astpkg.FuncDecl); ok {
            for _, p := range fd.Params {
                walk(p.Type)
            }
            for _, r := range fd.Result {
                walk(r)
            }
        }
    }
    return diags
}

