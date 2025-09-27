package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// ImportAliases returns the set of in-scope import names usable for dotted refs.
// It prefers explicit alias; otherwise uses the last path segment of the import path.
func ImportAliases(f *ast.File) map[string]struct{} {
    out := map[string]struct{}{}
    if f == nil { return out }
    for _, d := range f.Decls {
        if im, ok := d.(*ast.ImportDecl); ok {
            alias := im.Alias
            if alias == "" {
                p := im.Path
                if i := lastSlash(p); i >= 0 && i+1 < len(p) { alias = p[i+1:] } else { alias = p }
            }
            if alias != "" { out[alias] = struct{}{} }
        }
    }
    return out
}

// lastSlash provided in resolution.go; reuse to keep behavior aligned.
