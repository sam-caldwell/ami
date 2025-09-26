package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "strings"
)

// analyzeSetTypes walks declared function signatures and struct fields to ensure
// any `set<T>` types meet minimal constraints: exactly one type argument, and
// element type T is not a pointer, slice, map, set, or slice, and has no generic args.
func analyzeSetTypes(f *astpkg.File) []diag.Diagnostic {
    var diags []diag.Diagnostic
    var walk func(t astpkg.TypeRef)
    walk = func(t astpkg.TypeRef) {
        if strings.ToLower(t.Name) == "set" {
            if len(t.Args) != 1 {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_SET_ARITY", Message: "set must have exactly one type argument: set<T>"})
            } else {
                e := t.Args[0]
                if e.Ptr || e.Slice {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_SET_ELEM_TYPE_INVALID", Message: "set element type cannot be pointer or slice"})
                }
                switch strings.ToLower(e.Name) {
                case "map", "set", "slice":
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_SET_ELEM_TYPE_INVALID", Message: "set element type cannot be map/set/slice"})
                }
                if len(e.Args) > 0 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_SET_ELEM_TYPE_INVALID", Message: "set element type cannot be generic"})
                }
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

