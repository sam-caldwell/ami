package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeEventTypeFlow performs a conservative type-flow check across pipeline edges.
// If both the upstream and downstream steps declare a `type(...)` attribute with a
// single string or ident argument, and they differ, emit E_EVENT_TYPE_FLOW at the
// downstream step position.
func AnalyzeEventTypeFlow(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        // collect step types and positions
        stepType := map[string]string{}
        stepPos := map[string]diag.Position{}
        for _, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok {
                stepPos[st.Name] = diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}
                for _, at := range st.Attrs {
                    if at.Name == "type" || at.Name == "Type" {
                        if len(at.Args) > 0 && at.Args[0].Text != "" {
                            stepType[st.Name] = at.Args[0].Text
                        }
                    }
                }
            }
        }
        // build edge list and indegree counts
        type edge struct{ from, to string }
        var edges []edge
        // compare across edges
        for _, s := range pd.Stmts {
            if e, ok := s.(*ast.EdgeStmt); ok {
                edges = append(edges, edge{from: e.From, to: e.To})
            }
        }
        // simple iterative propagation: forward then backward passes
        // forward: if all preds have same non-empty type, assign to node
        for pass := 0; pass < 2; pass++ {
            changed := false
            for _, e := range edges {
                // forward check for target if unset
                if stepType[e.to] == "" {
                    // gather preds of e.to
                    t := ""
                    consistent := true
                    for _, ee := range edges {
                        if ee.to == e.to {
                            pt := stepType[ee.from]
                            if pt == "" { consistent = false; break }
                            if t == "" { t = pt } else if t != pt { consistent = false; break }
                        }
                    }
                    if consistent && t != "" { stepType[e.to] = t; changed = true }
                }
                // backward: if source unset and all succs have same type
                if stepType[e.from] == "" {
                    t := ""
                    consistent := true
                    for _, ee := range edges {
                        if ee.from == e.from {
                            pt := stepType[ee.to]
                            if pt == "" { consistent = false; break }
                            if t == "" { t = pt } else if t != pt { consistent = false; break }
                        }
                    }
                    if consistent && t != "" { stepType[e.from] = t; changed = true }
                }
            }
            if !changed { break }
        }
        // emit mismatches where both sides have explicit/derived types
        for _, e := range edges {
            tFrom := stepType[e.from]
            tTo := stepType[e.to]
            if tFrom != "" && tTo != "" && tFrom != tTo {
                pos := stepPos[e.to]
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EVENT_TYPE_FLOW", Message: "event type mismatch across edge", Pos: &pos})
            }
        }
        // MultiPath inputs: if Collect has >1 typed upstream with differing types, flag mismatch at Collect
        // gather incoming types per node
        inTypes := map[string]map[string]struct{}{}
        for _, e := range edges {
            if stepType[e.from] != "" {
                if inTypes[e.to] == nil { inTypes[e.to] = map[string]struct{}{} }
                inTypes[e.to][stepType[e.from]] = struct{}{}
            }
        }
        for node, set := range inTypes {
            if len(set) > 1 {
                pos := stepPos[node]
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EVENT_TYPE_FLOW", Message: "event types differ across multiple upstreams", Pos: &pos})
            }
        }
    }
    return out
}
