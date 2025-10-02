package main

import (
    "strings"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

// applyJSONExcludes drops requested fields from the graph prior to encoding.
// Supported: "attrs" removes Edge.Attrs objects from all edges.
func applyJSONExcludes(g graph.Graph, excludes []string) graph.Graph {
    has := map[string]bool{}
    for _, e := range excludes { has[strings.ToLower(strings.TrimSpace(e))] = true }
    if has["attrs"] {
        for i := range g.Edges { g.Edges[i].Attrs = nil }
    }
    return g
}

