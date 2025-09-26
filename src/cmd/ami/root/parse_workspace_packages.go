package root

import (
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// parseWorkspacePackages extracts a map of package name -> root directory
// from the loosely-typed workspace.Packages field.
func parseWorkspacePackages(ws *workspace.Workspace) map[string]string {
    out := map[string]string{}
    for _, p := range ws.Packages {
        m, ok := p.(map[string]any)
        if !ok {
            continue
        }
        for name, v := range m {
            // value is expected to be a map with at least 'root'
            if vm, ok := v.(map[string]any); ok {
                if r, ok := vm["root"].(string); ok && strings.TrimSpace(r) != "" {
                    out[name] = r
                }
            }
        }
    }
    // If no packages parsed, default to main: ./src (common scaffold)
    if len(out) == 0 {
        out["main"] = "./src"
    }
    return out
}

// lintOrder returns packages in the order they should be linted.
func lintOrder(pkgRoots map[string]string) []string {
    // Helper to read a representative unit for a package to extract imports.
    importsFor := func(pkg string) []string {
        root, ok := pkgRoots[pkg]
        if !ok {
            return nil
        }
        // Prefer main.ami; otherwise first *.ami by name.
        mainPath := filepath.Join(root, "main.ami")
        var path string
        if fi, err := os.Stat(mainPath); err == nil && !fi.IsDir() {
            path = mainPath
        } else {
            list, _ := filepath.Glob(filepath.Join(root, "*.ami"))
            sort.Strings(list)
            if len(list) > 0 {
                path = list[0]
            }
        }
        if path == "" {
            return nil
        }
        b, err := os.ReadFile(path)
        if err != nil {
            return nil
        }
        return parser.ExtractImports(string(b))
    }

    // DFS
    visited := map[string]bool{}
    order := []string{}
    var visit func(string)
    visit = func(pkg string) {
        if visited[pkg] {
            return
        }
        visited[pkg] = true
        // For each import that is a workspace-local package, visit first
        for _, imp := range importsFor(pkg) {
            if _, ok := pkgRoots[imp]; ok {
                visit(imp)
            }
        }
        order = append(order, pkg)
    }
    // Always start from main if present
    if _, ok := pkgRoots["main"]; ok {
        visit("main")
    }
    // Include any remaining packages that weren't reachable from main in stable order
    var rest []string
    for k := range pkgRoots {
        if !visited[k] {
            rest = append(rest, k)
        }
    }
    sort.Strings(rest)
    for _, k := range rest {
        visit(k)
    }
    return order
}

