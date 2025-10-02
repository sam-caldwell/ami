package main

import (
    "strings"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

// graphContains checks if g's name or any node label/kind contains focus (lowercase substring).
func graphContains(g graph.Graph, focus string) bool {
    if strings.Contains(strings.ToLower(g.Name), focus) { return true }
    for _, n := range g.Nodes {
        if strings.Contains(strings.ToLower(n.Label), focus) || strings.Contains(strings.ToLower(n.Kind), focus) { return true }
    }
    return false
}

