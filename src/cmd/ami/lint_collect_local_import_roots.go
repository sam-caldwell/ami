package main

import "github.com/sam-caldwell/ami/src/ami/workspace"

// collectLocalImportRoots walks local (./...) imports recursively starting from pkg,
// returning a DFS order with children before parents (child-first) and duplicates eliminated.
func collectLocalImportRoots(ws *workspace.Workspace, pkg *workspace.Package) []string {
    visited := make(map[string]bool)
    var order []string
    var dfs func(p *workspace.Package)
    dfs = func(p *workspace.Package) {
        // Traverse each local import
        for _, entry := range p.Import {
            path, _ := splitImportConstraint(entry)
            if !stringsHasPrefixAny(path, []string{"./"}) || stringsHasPrefixAny(path, []string{"../"}) { continue }
            if visited[path] { continue }
            visited[path] = true
            if child := findPackageByRoot(ws, path); child != nil { dfs(child) }
            order = append(order, path)
        }
    }
    dfs(pkg)
    return order
}

