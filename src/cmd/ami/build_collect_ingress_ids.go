package main

import (
    "path/filepath"
    "sort"
    "os"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// collectIngressIDs scans the workspace packages and returns a stable list of ingress identifiers,
// formatted as "<pkg>.<pipeline>". It parses .ami files under each package root.
func collectIngressIDs(ws workspace.Workspace, root string) []string {
    var result []string
    for _, entry := range ws.Packages {
        pkg := entry.Package
        if pkg.Root == "" || pkg.Name == "" { continue }
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
            if af == nil || af.PackageName == "" { continue }
            for _, d := range af.Decls {
                if pd, ok := d.(*ast.PipelineDecl); ok {
                    if pd.Name != "" { result = append(result, af.PackageName+"."+pd.Name) }
                }
            }
        }
    }
    sort.Strings(result)
    return result
}

