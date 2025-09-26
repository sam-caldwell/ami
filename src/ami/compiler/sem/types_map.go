package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "strings"
)

// analyzeMapTypes walks declared function signatures and struct fields to ensure
// any `map<K,V>` types meet minimal constraints: exactly two type arguments, and
// key type K is not a pointer, slice, map, set, or slice, and has no generic args.
func analyzeMapTypes(f *astpkg.File) []diag.Diagnostic {
    var diags []diag.Diagnostic
    var walk func(t astpkg.TypeRef)
    walk = func(t astpkg.TypeRef) {
        if strings.ToLower(t.Name) == "map" {
            if len(t.Args) != 2 {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MAP_ARITY", Message: "map must have exactly two type arguments: map<K,V>"})
            } else {
                k := t.Args[0]
                if k.Ptr || k.Slice {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MAP_KEY_TYPE_INVALID", Message: "map key type cannot be pointer or slice"})
                }
                switch strings.ToLower(k.Name) {
                case "map", "set", "slice":
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MAP_KEY_TYPE_INVALID", Message: "map key type cannot be map/set/slice"})
                }
                if len(k.Args) > 0 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MAP_KEY_TYPE_INVALID", Message: "map key type cannot be generic"})
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

