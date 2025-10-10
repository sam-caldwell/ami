package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// writePipelinesDebug writes pipelines debug JSON for a parsed file.
func writePipelinesDebug(pkg, unit string, f *ast.File) (string, error) {
    var entries []pipelineEntry
    defaultDelivery := "atLeastOnce"
    if f != nil {
        for _, pr := range f.Pragmas {
            if pr.Domain == "backpressure" {
                if pol, ok := pr.Params["policy"]; ok {
                    switch pol {
                    case "dropOldest", "dropNewest": defaultDelivery = "bestEffort"
                    case "block": defaultDelivery = "atLeastOnce"
                    }
                }
            }
        }
    }
    var conc *pipeConcurrency
    if f != nil {
        for _, pr := range f.Pragmas {
            if pr.Domain != "concurrency" { continue }
            if conc == nil { conc = &pipeConcurrency{} }
            if pr.Key == "workers" {
                if v, ok := pr.Params["count"]; ok && v != "" { if n := getInt(v); n > 0 { conc.Workers = n } } else if pr.Value != "" { if n := getInt(pr.Value); n > 0 { conc.Workers = n } }
            }
            if pr.Key == "schedule" && pr.Value != "" { conc.Schedule = pr.Value }
            if pr.Key == "limits" {
                if conc.Limits == nil { conc.Limits = map[string]int{} }
                for k, v := range pr.Params { if n := getInt(v); n > 0 { conc.Limits[k] = n } }
                for _, a := range pr.Args { if eq := strings.IndexByte(a, '='); eq > 0 { k := a[:eq]; v := a[eq+1:]; if n := getInt(v); n > 0 { conc.Limits[k] = n } } }
            }
        }
    }
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        var steps []pipelineOp
        var edges []pipeEdge
        occs := map[string][]occ{}
        nameCount := map[string]int{}
        for si, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok { nameCount[st.Name]++; occs[st.Name] = append(occs[st.Name], occ{id: nameCount[st.Name], stmtIdx: si}) }
        }
        // inline worker counter for deterministic synthetic names
        inlineIdx := 0
        for _, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok {
                var args []string
                // detect inline worker literal and normalize to a generated name
                genInlineName := func(id int) string {
                    // deterministic synthetic name per unit/pipeline/occurrence
                    name := "InlineWorker_" + unit + "_" + pd.Name + "_" + itoa(id)
                    // keep as plain identifier; symbol sanitization performed later
                    return name
                }
                // copy args; rewrite worker literal if present
                // prefer named worker=...; else first positional
                workerArgIdx := -1
                workerIsNamed := false
                for i, a := range st.Args {
                    s := a.Text
                    if eq := strings.IndexByte(s, '='); eq > 0 {
                        key := strings.TrimSpace(s[:eq])
                        if strings.EqualFold(key, "worker") {
                            workerArgIdx = i
                            workerIsNamed = true
                        }
                    } else if i == 0 {
                        workerArgIdx = 0
                    }
                }
                // build initial args
                for _, a := range st.Args { args = append(args, a.Text) }
                // normalize inline function literal
                if strings.EqualFold(st.Name, "Transform") && workerArgIdx >= 0 {
                    text := args[workerArgIdx]
                    rhs := text
                    if workerIsNamed {
                        if eq := strings.IndexByte(text, '='); eq > 0 { rhs = strings.TrimSpace(text[eq+1:]) }
                    }
                    if strings.HasPrefix(strings.TrimSpace(rhs), "func") {
                        inlineIdx++
                        wname := genInlineName(inlineIdx)
                        if workerIsNamed {
                            args[workerArgIdx] = "worker=" + wname
                        } else {
                            args[workerArgIdx] = wname
                        }
                    }
                }
                id := 0
                if arr := occs[st.Name]; len(arr) > 0 { for _, o := range arr { if o.stmtIdx == indexOfStmt(pd.Stmts, s) { id = o.id; break } } }
                op := pipelineOp{Name: st.Name, ID: id, Args: args}
                op.Edge = &edgeAttrs{Bounded: false, Delivery: defaultDelivery}
                var rawAttrs []pipeAttr
                var norm pipeMergeNorm
                hadMerge := false
                for _, at := range st.Attrs {
                    var aargs []string
                    for _, aa := range at.Args { aargs = append(aargs, aa.Text) }
                    rawAttrs = append(rawAttrs, pipeAttr{Name: at.Name, Args: aargs})
                    if (at.Name == "type" || at.Name == "Type") && len(aargs) > 0 && aargs[0] != "" { op.Edge.Type = aargs[0] }
                    if len(at.Name) >= 6 && at.Name[:6] == "merge." {
                        var margs []string
                        for _, aa := range at.Args { margs = append(margs, aa.Text) }
                        op.Merge = append(op.Merge, pipeMergeAttr{Name: at.Name, Args: margs})
                        hadMerge = true
                        if at.Name == "merge.Buffer" {
                            kv := map[string]string{}
                            for _, a := range margs { if eq := strings.IndexByte(a, '='); eq > 0 { kv[strings.ToLower(strings.TrimSpace(a[:eq]))] = strings.TrimSpace(a[eq+1:]) } }
                            capVal := 0
                            pol := ""
                            if v := kv["capacity"]; v != "" { capVal = getInt(v) } else if len(margs) > 0 { capVal = getInt(margs[0]) }
                            if v := kv["policy"]; v != "" { pol = v } else if len(margs) > 1 { pol = margs[1] }
                            if capVal > 0 { op.Edge.Bounded = true }
                            if pol == "dropOldest" || pol == "dropNewest" { op.Edge.Delivery = "bestEffort" }
                            if pol == "block" { op.Edge.Delivery = "atLeastOnce" }
                            nb := struct{ Capacity int `json:"capacity"`; Policy string `json:"policy,omitempty"` }{Capacity: capVal, Policy: pol}
                            norm.Buffer = &nb
                        }
                        if at.Name == "merge.Stable" { norm.Stable = true }
                        if at.Name == "merge.Sort" {
                            var field, order string
                            if len(margs) > 0 { field = margs[0] }
                            if len(margs) > 1 { order = margs[1] } else { order = "asc" }
                            norm.Sort = append(norm.Sort, struct{ Field string `json:"field"`; Order string `json:"order"` }{Field: field, Order: order})
                        }
                        if at.Name == "merge.Key" { if len(margs) > 0 { norm.Key = margs[0] } }
                        if at.Name == "merge.PartitionBy" { if len(margs) > 0 { norm.PartitionBy = margs[0] } }
                        if at.Name == "merge.Timeout" {
                            kv := map[string]string{}
                            for _, a := range margs { if eq := strings.IndexByte(a, '='); eq > 0 { kv[strings.ToLower(strings.TrimSpace(a[:eq]))] = strings.TrimSpace(a[eq+1:]) } }
                            v := ""
                            if kv["ms"] != "" { v = kv["ms"] } else if len(margs) > 0 { v = margs[0] }
                            to := 0
                            for i := 0; i < len(v); i++ { if v[i] >= '0' && v[i] <= '9' { to = to*10 + int(v[i]-'0') } else { to = 0; break } }
                            norm.TimeoutMs = to
                        }
                        if at.Name == "merge.Window" {
                            kv := map[string]string{}
                            for _, a := range margs { if eq := strings.IndexByte(a, '='); eq > 0 { kv[strings.ToLower(strings.TrimSpace(a[:eq]))] = strings.TrimSpace(a[eq+1:]) } }
                            v := ""
                            if kv["size"] != "" { v = kv["size"] } else if len(margs) > 0 { v = margs[0] }
                            w := 0
                            for i := 0; i < len(v); i++ { if v[i] >= '0' && v[i] <= '9' { w = w*10 + int(v[i]-'0') } else { w = 0; break } }
                            norm.Window = w
                        }
                        if at.Name == "merge.Watermark" {
                            kv := map[string]string{}
                            for _, a := range margs { if eq := strings.IndexByte(a, '='); eq > 0 { kv[strings.ToLower(strings.TrimSpace(a[:eq]))] = strings.TrimSpace(a[eq+1:]) } }
                            var field, late string
                            if kv["field"] != "" { field = kv["field"] } else if len(margs) > 0 { field = margs[0] }
                            if kv["lateness"] != "" { late = kv["lateness"] } else if len(margs) > 1 { late = margs[1] }
                            wm := struct{ Field string `json:"field"`; Lateness string `json:"lateness"` }{Field: field, Lateness: late}
                            norm.Watermark = &wm
                        }
                        if at.Name == "merge.Dedup" { if len(margs) > 0 { norm.Dedup = margs[0] } else { norm.Dedup = "" } }
                    }
                    if (at.Name == "edge.MultiPath" || at.Name == "MultiPath") && st.Name == "Collect" {
                        var margs []string
                        for _, aa := range at.Args { margs = append(margs, aa.Text) }
                        var inputs []string
                        for _, s2 := range pd.Stmts { if e, ok := s2.(*ast.EdgeStmt); ok && e.To == st.Name { inputs = append(inputs, e.From) } }
                        sort.Strings(inputs)
                        op.MultiPath = &pipeMultiPath{Args: margs, Inputs: inputs}
                    }
                }
                if hadMerge {
                    // Attach normalized merge view when any merge.* attribute present
                    // Tests assert presence and content under mergeNorm
                    op.MergeNorm = &norm
                }
                // Always snapshot multipath inputs for Collect, even without explicit attribute
                if op.MultiPath == nil && st.Name == "Collect" {
                    var inputs []string
                    for _, s2 := range pd.Stmts {
                        if e2, ok2 := s2.(*ast.EdgeStmt); ok2 && e2.To == st.Name {
                            // attribute this edge to the nearest following Collect occurrence
                            toID := 0
                            if arr := occs[e2.To]; len(arr) > 0 { toID = nearestOccAfterOcc(arr, indexOfStmt(pd.Stmts, s2)) }
                            if toID == id { inputs = append(inputs, e2.From) }
                        }
                    }
                    sort.Strings(inputs)
                    op.MultiPath = &pipeMultiPath{Inputs: inputs}
                }
                op.Attrs = rawAttrs
                steps = append(steps, op)
            } else if e, ok := s.(*ast.EdgeStmt); ok {
                id := 0
                if arr := occs[e.From]; len(arr) > 0 { id = nearestOccBeforeOcc(arr, indexOfStmt(pd.Stmts, s)) }
                toID := 0
                if arr := occs[e.To]; len(arr) > 0 { toID = nearestOccAfterOcc(arr, indexOfStmt(pd.Stmts, s)) }
                edges = append(edges, pipeEdge{From: e.From, To: e.To, FromID: id, ToID: toID})
            }
        }
        var conn *pipeConn
        if len(edges) > 0 {
            nodes := map[string]bool{}
            adj := map[string][]string{}
            radj := map[string][]string{}
            for _, s := range steps { nodes[s.Name] = true }
            for _, e := range edges { nodes[e.From] = true; nodes[e.To] = true; adj[e.From] = append(adj[e.From], e.To); radj[e.To] = append(radj[e.To], e.From) }
            reach := map[string]bool{}
            var stack []string
            stack = append(stack, "ingress")
            for len(stack) > 0 {
                n := stack[len(stack)-1]; stack = stack[:len(stack)-1]
                if reach[n] { continue }
                reach[n] = true
                for _, m := range adj[n] { if !reach[m] { stack = append(stack, m) } }
            }
            var disc []string
            for n := range nodes { if n != "ingress" && n != "egress" && !reach[n] { disc = append(disc, n) } }
            sort.Strings(disc)
            reachToEgress := map[string]bool{}
            var rstack []string
            rstack = append(rstack, "egress")
            for len(rstack) > 0 {
                n := rstack[len(rstack)-1]; rstack = rstack[:len(rstack)-1]
                if reachToEgress[n] { continue }
                reachToEgress[n] = true
                for _, m := range radj[n] { if !reachToEgress[m] { rstack = append(rstack, m) } }
            }
            var notFromIngress []string
            var cannotToEgress []string
            for _, st := range steps { if !reach[st.Name] { notFromIngress = append(notFromIngress, st.Name) }; if !reachToEgress[st.Name] { cannotToEgress = append(cannotToEgress, st.Name) } }
            sort.Strings(notFromIngress); sort.Strings(cannotToEgress)
            occNodes := map[pipeNodeRef]struct{}{}
            for _, st := range steps { if st.ID > 0 { occNodes[pipeNodeRef{Name: st.Name, ID: st.ID}] = struct{}{} } }
            oadj := map[pipeNodeRef][]pipeNodeRef{}
            oradj := map[pipeNodeRef][]pipeNodeRef{}
            for _, e := range edges { if e.FromID > 0 && e.ToID > 0 { a := pipeNodeRef{Name: e.From, ID: e.FromID}; b := pipeNodeRef{Name: e.To, ID: e.ToID}; oadj[a] = append(oadj[a], b); oradj[b] = append(oradj[b], a) } }
            var ingressOcc *pipeNodeRef
            var egressOccs []pipeNodeRef
            for n := range occNodes { if strings.ToLower(n.Name) == "ingress" { x := n; ingressOcc = &x }; if strings.ToLower(n.Name) == "egress" { egressOccs = append(egressOccs, n) } }
            var notFromIngressIDs []pipeNodeRef
            var cannotToEgressIDs []pipeNodeRef
            if ingressOcc != nil {
                vis := map[pipeNodeRef]bool{}
                st := []pipeNodeRef{*ingressOcc}
                for len(st) > 0 {
                    n := st[len(st)-1]; st = st[:len(st)-1]
                    if vis[n] { continue }
                    vis[n] = true
                    for _, m := range oadj[n] { if !vis[m] { st = append(st, m) } }
                }
                for n := range occNodes { if !vis[n] { notFromIngressIDs = append(notFromIngressIDs, n) } }
            }
            if len(egressOccs) > 0 {
                vis := map[pipeNodeRef]bool{}
                st := append([]pipeNodeRef(nil), egressOccs...)
                for len(st) > 0 {
                    n := st[len(st)-1]; st = st[:len(st)-1]
                    if vis[n] { continue }
                    vis[n] = true
                    for _, m := range oradj[n] { if !vis[m] { st = append(st, m) } }
                }
                for n := range occNodes { if !vis[n] { cannotToEgressIDs = append(cannotToEgressIDs, n) } }
            }
            sort.SliceStable(notFromIngressIDs, func(i, j int) bool { if notFromIngressIDs[i].Name == notFromIngressIDs[j].Name { return notFromIngressIDs[i].ID < notFromIngressIDs[j].ID }; return notFromIngressIDs[i].Name < notFromIngressIDs[j].Name })
            sort.SliceStable(cannotToEgressIDs, func(i, j int) bool { if cannotToEgressIDs[i].Name == cannotToEgressIDs[j].Name { return cannotToEgressIDs[i].ID < cannotToEgressIDs[j].ID }; return cannotToEgressIDs[i].Name < cannotToEgressIDs[j].Name })
            conn = &pipeConn{HasEdges: true, IngressToEgress: reach["egress"], Disconnected: disc, UnreachableFromIngress: notFromIngress, CannotReachEgress: cannotToEgress, UnreachableFromIngressIDs: notFromIngressIDs, CannotReachEgressIDs: cannotToEgressIDs}
        }
        if len(edges) > 0 {
            sort.SliceStable(edges, func(i, j int) bool { if edges[i].From == edges[j].From { return edges[i].To < edges[j].To }; return edges[i].From < edges[j].From })
        }
        entries = append(entries, pipelineEntry{Name: pd.Name, Steps: steps, Edges: edges, Conn: conn})
    }
    if len(entries) == 0 { entries = []pipelineEntry{{Name: "", Steps: []pipelineOp{{Name: "", Edge: &edgeAttrs{Bounded: false, Delivery: defaultDelivery}}}}} }
    sort.SliceStable(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
    obj := pipelineList{Schema: "pipelines.v1", Package: pkg, Unit: unit, Concurrency: conc, Pipelines: entries}
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(obj, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, unit+".pipelines.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}
