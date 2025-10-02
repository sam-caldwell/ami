package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// collectEdges returns all edge entries for a parsed file, tagged with unit.
func collectEdges(unit string, f *ast.File) []edgeEntry {
    var out []edgeEntry
    if f == nil { return out }
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        // build occurrences map to compute IDs
        type occ struct{ id int; idx int }
        occs := map[string][]occ{}
        count := map[string]int{}
        for i, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok {
                count[st.Name]++
                occs[st.Name] = append(occs[st.Name], occ{id: count[st.Name], idx: i})
            }
        }
        // build a quick index of step attrs by step name
        stepAttrs := map[string][]ast.Attr{}
        for _, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok {
                stepAttrs[st.Name] = st.Attrs
            }
        }
        // Build connectivity (name-level) for hints
        nodes := map[string]bool{}
        adj := map[string][]string{}
        radj := map[string][]string{}
        for _, s := range pd.Stmts {
            switch n := s.(type) {
            case *ast.StepStmt:
                nodes[n.Name] = true
            case *ast.EdgeStmt:
                nodes[n.From] = true; nodes[n.To] = true
                adj[n.From] = append(adj[n.From], n.To)
                radj[n.To] = append(radj[n.To], n.From)
            }
        }
        // Reachability from ingress
        reach := map[string]bool{}
        var stReach []string
        stReach = append(stReach, "ingress")
        for len(stReach) > 0 {
            n := stReach[len(stReach)-1]
            stReach = stReach[:len(stReach)-1]
            if reach[n] { continue }
            reach[n] = true
            for _, m := range adj[n] { if !reach[m] { stReach = append(stReach, m) } }
        }
        // Reachability to egress
        reachTo := map[string]bool{}
        var stTo []string
        stTo = append(stTo, "egress")
        for len(stTo) > 0 {
            n := stTo[len(stTo)-1]
            stTo = stTo[:len(stTo)-1]
            if reachTo[n] { continue }
            reachTo[n] = true
            for _, m := range radj[n] { if !reachTo[m] { stTo = append(stTo, m) } }
        }

        for i, s := range pd.Stmts {
            if e, ok := s.(*ast.EdgeStmt); ok {
                // derive bounded/delivery from target step attributes
                bounded := false
                delivery := "atLeastOnce"
                etype := ""
                tiny := false
                for _, at := range stepAttrs[e.To] {
                    if at.Name == "dropOldest" || at.Name == "dropNewest" { delivery = "bestEffort" }
                    if at.Name == "merge.Buffer" {
                        if len(at.Args) > 0 {
                            if at.Args[0].Text != "0" && at.Args[0].Text != "" { bounded = true }
                            if at.Args[0].Text == "0" || at.Args[0].Text == "1" {
                                if len(at.Args) > 1 {
                                    pol := at.Args[1].Text
                                    if pol == "dropOldest" || pol == "dropNewest" { tiny = true }
                                }
                            }
                        }
                        if len(at.Args) > 1 {
                            pol := at.Args[1].Text
                            if pol == "dropOldest" || pol == "dropNewest" { delivery = "bestEffort" }
                            if pol == "block" { delivery = "atLeastOnce" }
                        }
                    }
                    if (at.Name == "type" || at.Name == "Type") && len(at.Args) > 0 {
                        if at.Args[0].Text != "" { etype = at.Args[0].Text }
                    }
                }
                // resolve IDs: nearest occurrence before for From; nearest at/after for To
                fromID := 0
                toID := 0
                if arr := occs[e.From]; len(arr) > 0 {
                    bestIdx := -1
                    for _, o := range arr { if o.idx <= i && o.idx >= bestIdx { bestIdx = o.idx; fromID = o.id } }
                }
                if arr := occs[e.To]; len(arr) > 0 {
                    bestIdx := 1<<30
                    for _, o := range arr { if o.idx >= i && o.idx <= bestIdx { bestIdx = o.idx; toID = o.id } }
                }
                fr := reach[e.From]
                tr := reachTo[e.To]
                on := fr && tr
                out = append(out, edgeEntry{Unit: unit, Pipeline: pd.Name, From: e.From, To: e.To, FromID: fromID, ToID: toID, Bounded: bounded, Delivery: delivery, Type: etype, Tiny: tiny, FromReachable: fr, ToReachable: tr, OnPath: on})
            }
        }
    }
    return out
}

