package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "strings"
)

// analyzeCycles builds a graph of pipeline→pipeline references via edge.Pipeline
// and emits E_CYCLE_DETECTED when a cycle is present unless a `#pragma cycle allow`
// directive is set at file level.
func analyzeCycles(f *astpkg.File) []diag.Diagnostic {
    var diags []diag.Diagnostic
    allow := false
    for _, d := range f.Directives {
        if strings.ToLower(d.Name) == "cycle" && strings.Contains(strings.ToLower(d.Payload), "allow") {
            allow = true
            break
        }
    }
    if allow {
        return diags
    }
    // collect pipelines and edges
    names := []string{}
    idx := map[string]int{}
    for _, n := range f.Decls {
        if p, ok := n.(astpkg.PipelineDecl); ok {
            idx[p.Name] = len(names)
            names = append(names, p.Name)
        }
    }
    g := make([][]int, len(names))
    addEdge := func(from string, to string) {
        i, ok1 := idx[from]
        j, ok2 := idx[to]
        if ok1 && ok2 {
            g[i] = append(g[i], j)
        }
    }
    for _, n := range f.Decls {
        p, ok := n.(astpkg.PipelineDecl)
        if !ok {
            continue
        }
        scan := func(args []string) {
            if spec, ok := parseEdgeSpecFromArgs(args); ok {
                if v, ok2 := spec.(pipeSpec); ok2 {
                    if v.Name != "" {
                        addEdge(p.Name, v.Name)
                    }
                }
            }
        }
        for _, st := range p.Steps {
            scan(st.Args)
        }
        for _, st := range p.ErrorSteps {
            scan(st.Args)
        }
    }
    // detect cycles via DFS
    visited := make([]int, len(names)) // 0=unvisited,1=visiting,2=done
    var dfs func(int) bool
    dfs = func(u int) bool {
        visited[u] = 1
        for _, v := range g[u] {
            if visited[v] == 1 {
                return true
            }
            if visited[v] == 0 && dfs(v) {
                return true
            }
        }
        visited[u] = 2
        return false
    }
    for i := range names {
        if visited[i] == 0 && dfs(i) {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_CYCLE_DETECTED", Message: "pipeline graph contains a cycle; add `#pragma cycle allow` with anti-deadlock strategy to permit"})
            break
        }
    }
    return diags
}

