package ascii

import (
    "sort"
    "strings"

    "github.com/sam-caldwell/ami/src/schemas/graph"
)

// RenderLine returns a minimal one-line ASCII representation of a pipeline graph.
// It is a scaffold implementation: it supports straight-line pipelines (ingress→...→egress)
// and falls back to a simple ID-sorted sequence when the graph is not a single chain.
func RenderLine(g graph.Graph, _ Options) string {
    // Build adjacency and in-degree maps
    next := make(map[string]string, len(g.Edges))
    indeg := make(map[string]int, len(g.Nodes))
    ids := make(map[string]graph.Node, len(g.Nodes))
    for _, n := range g.Nodes { ids[n.ID] = n }
    for _, e := range g.Edges {
        // Only keep first outgoing edge per node to model a chain; ignore branches for now
        if _, ok := next[e.From]; !ok {
            next[e.From] = e.To
        }
        indeg[e.To]++
        // ensure keys exist in indeg
        if _, ok := indeg[e.From]; !ok { indeg[e.From] = 0 }
    }
    // Find a start node (in-degree 0)
    start := ""
    for id := range ids {
        if indeg[id] == 0 {
            start = id
            break
        }
    }
    var order []string
    if start != "" && len(g.Edges) > 0 {
        // Walk the chain
        cur := start
        seen := make(map[string]struct{}, len(ids))
        for cur != "" {
            if _, ok := ids[cur]; !ok { break }
            if _, ok := seen[cur]; ok { break }
            seen[cur] = struct{}{}
            order = append(order, cur)
            cur = next[cur]
        }
    }
    if len(order) == 0 || len(order) < len(ids) {
        // Fallback: stable ID ordering
        for id := range ids { order = append(order, id) }
        sort.Strings(order)
    }
    // Format nodes
    parts := make([]string, 0, len(order))
    for _, id := range order {
        n := ids[id]
        lbl := n.Label
        if strings.TrimSpace(lbl) == "" { lbl = n.Kind }
        switch strings.ToLower(n.Kind) {
        case "ingress", "egress":
            parts = append(parts, "["+lbl+"]")
        default:
            parts = append(parts, "("+lbl+")")
        }
    }
    return strings.Join(parts, " --> ")
}

