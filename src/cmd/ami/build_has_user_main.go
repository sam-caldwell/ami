package main

import (
    "os"
    "path/filepath"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// hasUserMain returns true if a package 'main' defines a function named 'main'.
func hasUserMain(ws workspace.Workspace, root string) bool {
    for _, entry := range ws.Packages {
        pkg := entry.Package
        if pkg.Root == "" || pkg.Name != "main" { continue }
        pdir := filepath.Clean(filepath.Join(root, pkg.Root))
        var files []string
        _ = filepath.WalkDir(pdir, func(path string, d os.DirEntry, err error) error {
            if err != nil || d.IsDir() { return nil }
            if filepath.Ext(path) == ".ami" { files = append(files, path) }
            return nil
        })
        for _, f := range files {
            b, err := os.ReadFile(f); if err != nil { continue }
            sf := &source.File{Name: f, Content: string(b)}
            af, _ := parser.New(sf).ParseFile()
            if af == nil { continue }
            for _, d := range af.Decls {
                if fn, ok := d.(*ast.FuncDecl); ok {
                    if fn.Name == "main" { return true }
                }
            }
        }
    }
    return false
}

