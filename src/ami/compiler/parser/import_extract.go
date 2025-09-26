package parser

import (
    "strings"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// ExtractImports finds import paths in a minimal Go-like syntax.
func ExtractImports(src string) []string {
    p := New(src)
    f := p.ParseFile()
    out := make([]string, len(f.Imports))
    copy(out, f.Imports)
    return out
}

// ImportItem describes an import with optional alias/constraint from source.
type ImportItem struct {
    Path       string
    Alias      string
    Constraint string
}

// ExtractImportItems returns detailed import items with optional constraints.
func ExtractImportItems(src string) []ImportItem {
    p := New(src)
    f := p.ParseFile()
    var out []ImportItem
    for _, d := range f.Decls {
        if id, ok := d.(astpkg.ImportDecl); ok {
            out = append(out, ImportItem{Path: id.Path, Alias: id.Alias, Constraint: id.Constraint})
        }
    }
    return out
}

func unquote(s string) string {
    if len(s) >= 2 && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
        return s[1 : len(s)-1]
    }
    return s
}

