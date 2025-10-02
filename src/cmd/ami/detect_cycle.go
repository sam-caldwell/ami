package main

import (
    "sort"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

// detectCycle performs a simple cycle check; returns true and a list of involved node IDs when a cycle exists.
func detectCycle(g graph.Graph) (bool, []string) {
    indeg := map[string]int{}
    adj := map[string][]string{}
    for _, n := range g.Nodes { indeg[n.ID] = 0 }
    for _, e := range g.Edges {
        adj[e.From] = append(adj[e.From], e.To)
        indeg[e.To]++
        if _, ok := indeg[e.From]; !ok { indeg[e.From] = 0 }
    }
    var q []string
    for id, d := range indeg { if d == 0 { q = append(q, id) } }
    processed := 0
    for len(q) > 0 {
        n := q[0]; q = q[1:]
        processed++
        for _, m := range adj[n] {
            indeg[m]--
            if indeg[m] == 0 { q = append(q, m) }
        }
    }
    if processed == len(indeg) { return false, nil }
    var cyc []string
    for id, d := range indeg { if d > 0 { cyc = append(cyc, id) } }
    sort.Strings(cyc)
    return true, cyc
}

