package ascii

import "github.com/sam-caldwell/ami/src/schemas/graph"

// findEdge returns the edge matching from→to if present.
func findEdge(g graph.Graph, from, to string) (graph.Edge, bool) {
    for _, e := range g.Edges { if e.From == from && e.To == to { return e, true } }
    return graph.Edge{}, false
}

