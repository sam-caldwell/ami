package main

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/workspace"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// lintImportCycles detects circular references among workspace-local imports (./foo style).
// Emits E_IMPORT_CYCLE for each detected cycle and returns error-level diagnostics.
func lintImportCycles(ws *workspace.Workspace) []diag.Record {
    var out []diag.Record
    if ws == nil { return out }
    now := time.Now().UTC()
    // Build adjacency using package roots (e.g., ./src, ./lib)
    adj := map[string][]string{}
    roots := map[string]bool{}
    for i := range ws.Packages {
        p := &ws.Packages[i].Package
        if p.Root == "" { continue }
        roots[p.Root] = true
        for _, entry := range p.Import {
            path, _ := splitImportConstraint(entry)
            if stringsHasPrefixAny(path, []string{"./"}) && !stringsHasPrefixAny(path, []string{"../"}) {
                if findPackageByRoot(ws, path) != nil {
                    adj[p.Root] = append(adj[p.Root], path)
                }
            }
        }
    }
    // DFS cycle detection across all roots
    const (
        white = 0
        gray  = 1
        black = 2
    )
    color := map[string]int{}
    stack := []string{}
    seenCycle := map[string]bool{}
    var emitCycle = func(cycle []string) {
        if len(cycle) == 0 { return }
        // canonicalize by rotating to lexicographically smallest node
        idx := 0
        for i := 1; i < len(cycle); i++ {
            if cycle[i] < cycle[idx] { idx = i }
        }
        canon := make([]string, 0, len(cycle))
        for i := 0; i < len(cycle); i++ { canon = append(canon, cycle[(idx+i)%len(cycle)]) }
        key := ""
        for _, n := range canon { key += n + "->" }
        if seenCycle[key] { return }
        seenCycle[key] = true
        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_IMPORT_CYCLE", Message: "circular local import detected", File: "ami.workspace", Data: map[string]any{"cycle": canon}})
    }
    var dfs func(u string)
    dfs = func(u string) {
        color[u] = gray
        stack = append(stack, u)
        for _, v := range adj[u] {
            if color[v] == white {
                dfs(v)
            } else if color[v] == gray {
                // found a back-edge; extract cycle from stack
                // find v in stack
                start := -1
                for i := len(stack) - 1; i >= 0; i-- { if stack[i] == v { start = i; break } }
                if start >= 0 {
                    cyc := append([]string{}, stack[start:]...)
                    emitCycle(cyc)
                }
            }
        }
        // pop u
        stack = stack[:len(stack)-1]
        color[u] = black
    }
    for r := range roots {
        if color[r] == white {
            dfs(r)
        }
    }
    return out
}

