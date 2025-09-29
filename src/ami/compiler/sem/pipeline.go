package sem

import (
    "time"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzePipelineSemantics enforces basic pipeline invariants:
// - first step must be `ingress` (E_PIPELINE_START_INGRESS)
// - last step must be `egress` (E_PIPELINE_END_EGRESS)
// - unknown steps emit E_UNKNOWN_NODE (only `ingress` and `egress` are recognized)
func AnalyzePipelineSemantics(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // scan for pipeline decls
    // detect duplicate pipeline names within the file
    seenPipes := map[string]diag.Position{}
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        if prev, exists := seenPipes[pd.Name]; exists {
            p := diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DUP_PIPELINE", Message: "duplicate pipeline name: " + pd.Name, Pos: &p, Data: map[string]any{"previous": prev}})
        } else {
            seenPipes[pd.Name] = diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}
        }
        // collect step statements only
        var steps []*ast.StepStmt
        for _, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok {
                steps = append(steps, st)
            }
        }
        // start check
        if len(steps) == 0 {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_START_INGRESS", Message: "pipeline must start with 'ingress'", File: "", Pos: &diag.Position{Line: pd.Pos.Line, Column: pd.Pos.Column, Offset: pd.Pos.Offset}})
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_END_EGRESS", Message: "pipeline must end with 'egress'", File: "", Pos: &diag.Position{Line: pd.Pos.Line, Column: pd.Pos.Column, Offset: pd.Pos.Offset}})
            continue
        }
        if steps[0].Name != "ingress" {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_START_INGRESS", Message: "pipeline must start with 'ingress'", File: "", Pos: &diag.Position{Line: steps[0].Pos.Line, Column: steps[0].Pos.Column, Offset: steps[0].Pos.Offset}})
        }
        // end check
        last := steps[len(steps)-1]
        if last.Name != "egress" {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_END_EGRESS", Message: "pipeline must end with 'egress'", File: "", Pos: &diag.Position{Line: last.Pos.Line, Column: last.Pos.Column, Offset: last.Pos.Offset}})
        }
        // position/uniqueness checks, io permissions, and unknown nodes
        ingressCount := 0
        egressCount := 0
        for _, st := range steps {
            if st.Name == "ingress" {
                if ingressCount > 0 { // duplicate beyond the first seen
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DUP_INGRESS", Message: "duplicate ingress", File: "", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
                }
                // position must be 0
                // check based on index by scanning again to find this st; simpler: compare to steps[0]
                if st != steps[0] {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_INGRESS_POSITION", Message: "'ingress' only allowed at position 0", File: "", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
                }
                ingressCount++
                continue
            }
            if st.Name == "egress" {
                if egressCount > 0 { // duplicate beyond the first seen
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DUP_EGRESS", Message: "duplicate egress", File: "", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
                }
                // position must be last
                if st != steps[len(steps)-1] {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EGRESS_POSITION", Message: "'egress' only allowed at last position", File: "", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
                }
                egressCount++
                continue
            }
            // IO capability detection: any non-ingress/egress node using io.*
            if strings.HasPrefix(st.Name, "io.") {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_IO_PERMISSION", Message: "io.* operations only allowed in ingress/egress nodes", File: "", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
            }
            // unknown nodes: treat core nodes as known (ingress/egress/transform/fanout/collect/mutable)
            nameLower := strings.ToLower(st.Name)
            if nameLower != "ingress" && nameLower != "egress" && nameLower != "transform" && nameLower != "fanout" && nameLower != "collect" && nameLower != "mutable" {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_UNKNOWN_NODE", Message: "unknown node: " + st.Name, File: "", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
            }
        }
        // Build a set of declared step names and positions for edge validation and connectivity.
        declared := map[string]struct{}{}
        stepPosMap := map[string]diag.Position{}
        for _, st := range steps { declared[st.Name] = struct{}{}; stepPosMap[st.Name] = diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset} }

        // Detect cycles using explicit edge statements (A -> B) and validate endpoints/directions.
        type edge struct{ from, to string; fromPos, toPos diag.Position }
        var edges []edge
        for _, s := range pd.Stmts {
            if e, ok := s.(*ast.EdgeStmt); ok {
                edges = append(edges, edge{
                    from:   e.From,
                    to:     e.To,
                    fromPos: diag.Position{Line: e.FromPos.Line, Column: e.FromPos.Column, Offset: e.FromPos.Offset},
                    toPos:   diag.Position{Line: e.ToPos.Line, Column: e.ToPos.Column, Offset: e.ToPos.Offset},
                })
            }
        }
        // Validate each edge endpoint references a declared step and enforce ingress/egress directionality.
        seenEdge := map[string]diag.Position{}
        for _, e := range edges {
            if _, ok := declared[e.from]; !ok {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGE_UNDECLARED_FROM", Message: "edge references undeclared node: " + e.from, Pos: &e.fromPos})
            }
            if _, ok := declared[e.to]; !ok {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGE_UNDECLARED_TO", Message: "edge references undeclared node: " + e.to, Pos: &e.toPos})
            }
            // Forbid inbound edges to ingress and outbound edges from egress.
            if strings.ToLower(e.to) == "ingress" {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGE_TO_INGRESS", Message: "edges cannot target 'ingress'", Pos: &e.toPos})
            }
            if strings.ToLower(e.from) == "egress" {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGE_FROM_EGRESS", Message: "edges cannot originate from 'egress'", Pos: &e.fromPos})
            }
            // Self-loop detection
            if e.from == e.to {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_SELF_EDGE", Message: "self edge is not allowed: " + e.from + " -> " + e.to, Pos: &e.fromPos})
            }
            // Duplicate edge detection
            key := e.from + "->" + e.to
            if prev, dup := seenEdge[key]; dup {
                // Report duplicate at second occurrence position
                p := e.fromPos
                if p.Line == 0 { p = prev }
                out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PIPELINE_DUP_EDGE", Message: "duplicate edge: " + key, Pos: &p})
            } else {
                seenEdge[key] = e.fromPos
            }
        }
        if len(edges) > 0 {
            // adjacency based on names; detect back-edge using DFS with recursion stack
            adj := map[string][]string{}
            radj := map[string][]string{}
            for _, e := range edges { adj[e.from] = append(adj[e.from], e.to); radj[e.to] = append(radj[e.to], e.from) }
            // helper to detect cycle
            visiting := map[string]bool{}
            visited := map[string]bool{}
            var hasCycle bool
            var dfs func(n string)
            dfs = func(n string) {
                if hasCycle || visiting[n] { hasCycle = true; return }
                if visited[n] { return }
                visiting[n] = true
                for _, m := range adj[n] { dfs(m) }
                visiting[n] = false
                visited[n] = true
            }
            for _, st := range steps { // iterate known nodes deterministically
                dfs(st.Name)
                if hasCycle { break }
            }
            if hasCycle {
                // Report a generic cycle error at the pipeline name position.
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_CYCLE", Message: "circular reference detected in pipeline '" + pd.Name + "'", Pos: &diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}})
            }

            // Connectivity checks: disconnected nodes and missing ingress→egress path.
            // Degree per node
            degree := map[string]int{}
            for _, e := range edges { degree[e.from]++; degree[e.to]++ }
            for _, st := range steps {
                if degree[st.Name] == 0 {
                    p := stepPosMap[st.Name]
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_NODE_DISCONNECTED", Message: "node has no incident edges: " + st.Name, Pos: &p})
                }
            }
            // Edge coverage: require ingress/egress presence when edges exist
            hasIngress := false
            hasEgress := false
            for _, st := range steps {
                nl := strings.ToLower(st.Name)
                if nl == "ingress" { hasIngress = true }
                if nl == "egress" { hasEgress = true }
            }
            if !hasIngress {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGES_WITHOUT_INGRESS", Message: "edges present but no 'ingress' declared", Pos: &diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}})
            }
            if !hasEgress {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGES_WITHOUT_EGRESS", Message: "edges present but no 'egress' declared", Pos: &diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}})
            }
            if hasIngress && hasEgress {
                // BFS/DFS from ingress using adj
                reach := map[string]bool{}
                var stack []string
                stack = append(stack, "ingress")
                for len(stack) > 0 {
                    n := stack[len(stack)-1]
                    stack = stack[:len(stack)-1]
                    if reach[n] { continue }
                    reach[n] = true
                    for _, m := range adj[n] { if !reach[m] { stack = append(stack, m) } }
                }
                if !reach["egress"] {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_NO_PATH_INGRESS_EGRESS", Message: "no path from ingress to egress", Pos: &diag.Position{Line: pd.NamePos.Line, Column: pd.NamePos.Column, Offset: pd.NamePos.Offset}})
                }
                // Reverse reachability to egress
                reachToEgress := map[string]bool{}
                var rstack []string
                rstack = append(rstack, "egress")
                for len(rstack) > 0 {
                    n := rstack[len(rstack)-1]
                    rstack = rstack[:len(rstack)-1]
                    if reachToEgress[n] { continue }
                    reachToEgress[n] = true
                    for _, m := range radj[n] { if !reachToEgress[m] { rstack = append(rstack, m) } }
                }
                // Flag unreachable and non-terminating nodes (exclude pure disconnected already handled)
                for _, st := range steps {
                    nameLower := strings.ToLower(st.Name)
                    if nameLower != "ingress" && !reach[st.Name] && degree[st.Name] > 0 {
                        p := stepPosMap[st.Name]
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_UNREACHABLE_FROM_INGRESS", Message: "node not reachable from ingress: " + st.Name, Pos: &p})
                    }
                    if nameLower != "egress" && !reachToEgress[st.Name] && degree[st.Name] > 0 {
                        p := stepPosMap[st.Name]
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_CANNOT_REACH_EGRESS", Message: "node cannot reach egress: " + st.Name, Pos: &p})
                    }
                }
            }
        }
    }
    return out
}
