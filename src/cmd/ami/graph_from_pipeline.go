package main

import (
    "fmt"
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

// graphFromPipeline constructs a simple straight-line graph from a PipelineDecl's step order.
func graphFromPipeline(pkg string, unit string, pd *ast.PipelineDecl) graph.Graph {
    g := graph.Graph{Package: pkg, Unit: unit, Name: pd.Name}
    // Collect steps and map names to node IDs
    var ids []string
    nameToID := map[string]string{}
    for i, s := range pd.Stmts {
        st, ok := s.(*ast.StepStmt); if !ok { continue }
        parts := strings.Split(st.Name, ".")
        base := parts[len(parts)-1]
        kind := strings.ToLower(base)
        label := base
        for _, at := range st.Attrs {
            if (at.Name == "type" || at.Name == "Type") && len(at.Args) > 0 {
                if at.Args[0].Text != "" { label = base + ":" + at.Args[0].Text }
            }
        }
        id := fmt.Sprintf("%02d:%s", i, strings.ToLower(base))
        g.Nodes = append(g.Nodes, graph.Node{ID: id, Kind: kind, Label: strings.ToLower(label)})
        ids = append(ids, id)
        if _, ok := nameToID[st.Name]; !ok { nameToID[st.Name] = id }
    }
    // Add explicit edges when present; otherwise chain sequentially
    var hasExplicit bool
    for _, s := range pd.Stmts { if _, ok := s.(*ast.EdgeStmt); ok { hasExplicit = true; break } }
    if hasExplicit {
        for _, s := range pd.Stmts {
            if e, ok := s.(*ast.EdgeStmt); ok {
                if fromID, ok1 := nameToID[e.From]; ok1 {
                    if toID, ok2 := nameToID[e.To]; ok2 {
                        attrs := deriveEdgeAttrs(pd, fromID, nameToID)
                        g.Edges = append(g.Edges, graph.Edge{From: fromID, To: toID, Attrs: attrs})
                    }
                }
            }
        }
    } else {
        for i := 0; i+1 < len(ids); i++ {
            attrs := deriveEdgeAttrs(pd, ids[i], nameToID)
            g.Edges = append(g.Edges, graph.Edge{From: ids[i], To: ids[i+1], Attrs: attrs})
        }
    }
    return g
}
