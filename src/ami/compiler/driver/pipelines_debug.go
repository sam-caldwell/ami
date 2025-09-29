package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

type pipelineList struct {
    Schema    string          `json:"schema"`
    Package   string          `json:"package"`
    Unit      string          `json:"unit"`
    Concurrency *pipeConcurrency `json:"concurrency,omitempty"`
    Pipelines []pipelineEntry `json:"pipelines"`
}
type pipelineEntry struct {
    Name  string       `json:"name"`
    Steps []pipelineOp `json:"steps"`
    Edges []pipeEdge   `json:"edges,omitempty"`
    Conn  *pipeConn    `json:"connectivity,omitempty"`
}
type pipelineOp struct {
    Name string   `json:"name"`
    ID   int      `json:"id,omitempty"`
    Args []string `json:"args,omitempty"`
    Edge *edgeAttrs `json:"edge,omitempty"`
    Merge []pipeMergeAttr `json:"merge,omitempty"`
    MergeNorm *pipeMergeNorm `json:"mergeNorm,omitempty"`
    MultiPath *pipeMultiPath `json:"multipath,omitempty"`
    Attrs []pipeAttr `json:"attrs,omitempty"`
}

type edgeAttrs struct {
    Bounded  bool   `json:"bounded"`
    Delivery string `json:"delivery"`
    Type     string `json:"type,omitempty"`
}

type pipeMergeAttr struct {
    Name string   `json:"name"`
    Args []string `json:"args"`
}

// Normalized merge config (scaffold)
type pipeMergeNorm struct {
    Buffer *struct{
        Capacity int    `json:"capacity"`
        Policy   string `json:"policy,omitempty"`
    } `json:"buffer,omitempty"`
    Stable bool `json:"stable,omitempty"`
    Sort   []struct{
        Field string `json:"field"`
        Order string `json:"order"`
    } `json:"sort,omitempty"`
    Key         string `json:"key,omitempty"`
    PartitionBy string `json:"partitionBy,omitempty"`
    TimeoutMs   int    `json:"timeoutMs,omitempty"`
    Window      int    `json:"window,omitempty"`
    Watermark   *struct{
        Field    string `json:"field"`
        Lateness string `json:"lateness"`
    } `json:"watermark,omitempty"`
    Dedup       string `json:"dedup,omitempty"`
}

type pipeMultiPath struct {
    Args   []string `json:"args"`
    Inputs []string `json:"inputs,omitempty"`
}

type pipeAttr struct {
    Name string   `json:"name"`
    Args []string `json:"args,omitempty"`
}

type pipeConcurrency struct {
    Workers  int    `json:"workers,omitempty"`
    Schedule string `json:"schedule,omitempty"`
    Limits   map[string]int `json:"limits,omitempty"`
}

type pipeEdge struct {
    From string `json:"from"`
    To   string `json:"to"`
    FromID int  `json:"fromId,omitempty"`
    ToID   int  `json:"toId,omitempty"`
}

type pipeConn struct {
    HasEdges         bool     `json:"hasEdges"`
    IngressToEgress  bool     `json:"ingressToEgress"`
    Disconnected     []string `json:"disconnected,omitempty"`
    UnreachableFromIngress []string `json:"unreachableFromIngress,omitempty"`
    CannotReachEgress      []string `json:"cannotReachEgress,omitempty"`
}

func getInt(s string) int {
    n := 0
    for i := 0; i < len(s); i++ {
        if s[i] >= '0' && s[i] <= '9' { n = n*10 + int(s[i]-'0') } else { return 0 }
    }
    return n
}

func indexOfStmt(stmts []ast.Stmt, target ast.Stmt) int {
    for i, s := range stmts { if s == target { return i } }
    return -1
}

func nearestOccBefore(arr []struct{ id int; stmtIdx int }, idx int) int {
    best := 0
    bestIdx := -1
    for _, o := range arr {
        if o.stmtIdx <= idx && o.stmtIdx >= bestIdx { bestIdx = o.stmtIdx; best = o.id }
    }
    return best
}

func nearestOccAfter(arr []struct{ id int; stmtIdx int }, idx int) int {
    best := 0
    bestIdx := 1<<30
    for _, o := range arr {
        if o.stmtIdx >= idx && o.stmtIdx <= bestIdx { bestIdx = o.stmtIdx; best = o.id }
    }
    return best
}

// writePipelinesDebug writes pipelines debug JSON for a parsed file.
func writePipelinesDebug(pkg, unit string, f *ast.File) (string, error) {
    var entries []pipelineEntry
    // derive defaults from pragmas
    defaultDelivery := "atLeastOnce"
    if f != nil {
        for _, pr := range f.Pragmas {
            if pr.Domain == "backpressure" {
                if pol, ok := pr.Params["policy"]; ok {
                    switch pol {
                    case "dropOldest", "dropNewest":
                        defaultDelivery = "bestEffort"
                    case "block":
                        defaultDelivery = "atLeastOnce"
                    }
                }
            }
        }
    }
    // Optional: read concurrency pragmas
    var conc *pipeConcurrency
    if f != nil {
        for _, pr := range f.Pragmas {
            if pr.Domain != "concurrency" { continue }
            if conc == nil { conc = &pipeConcurrency{} }
            if pr.Key == "workers" {
                if v, ok := pr.Params["count"]; ok && v != "" { if n := getInt(v); n > 0 { conc.Workers = n } } else if pr.Value != "" { if n := getInt(pr.Value); n > 0 { conc.Workers = n } }
            }
            if pr.Key == "schedule" { if pr.Value != "" { conc.Schedule = pr.Value } }
            if pr.Key == "limits" {
                if conc.Limits == nil { conc.Limits = map[string]int{} }
                for k, v := range pr.Params { if n := getInt(v); n > 0 { conc.Limits[k] = n } }
                for _, a := range pr.Args {
                    if eq := strings.IndexByte(a, '='); eq > 0 {
                        k := a[:eq]
                        v := a[eq+1:]
                        if n := getInt(v); n > 0 { conc.Limits[k] = n }
                    }
                }
            }
        }
    }

    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        var steps []pipelineOp
        var edges []pipeEdge
        // Track occurrences per step name with their statement index in pd.Stmts
        type occ struct{ id int; stmtIdx int }
        occs := map[string][]occ{}
        // Pre-scan to assign IDs per occurrence in order
        nameCount := map[string]int{}
        for si, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok {
                nameCount[st.Name]++
                occs[st.Name] = append(occs[st.Name], occ{id: nameCount[st.Name], stmtIdx: si})
            }
        }
        for _, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok {
                var args []string
                for _, a := range st.Args { args = append(args, a.Text) }
                // Determine this step's ID via occurrence index at this position
                id := 0
                if arr := occs[st.Name]; len(arr) > 0 {
                    // find matching occurrence by stmtIdx
                    for _, o := range arr { if o.stmtIdx == indexOfStmt(pd.Stmts, s) { id = o.id; break } }
                }
                op := pipelineOp{Name: st.Name, ID: id, Args: args}
                // default edge attributes (scaffold for #pragma backpressure)
                op.Edge = &edgeAttrs{Bounded: false, Delivery: defaultDelivery}
                // attributes
                var rawAttrs []pipeAttr
                var norm pipeMergeNorm
                for _, at := range st.Attrs {
                    // generic list
                    var aargs []string
                    for _, aa := range at.Args { aargs = append(aargs, aa.Text) }
                    rawAttrs = append(rawAttrs, pipeAttr{Name: at.Name, Args: aargs})
                    if (at.Name == "type" || at.Name == "Type") && len(aargs) > 0 && aargs[0] != "" {
                        op.Edge.Type = aargs[0]
                    }
                    // merge.* attributes captured verbatim (scaffold)
                    if len(at.Name) >= 6 && at.Name[:6] == "merge." {
                        var margs []string
                        for _, aa := range at.Args { margs = append(margs, aa.Text) }
                        op.Merge = append(op.Merge, pipeMergeAttr{Name: at.Name, Args: margs})
                        if at.Name == "merge.Buffer" {
                            // Support keyed args (capacity=, policy=) with last-write-wins.
                            kv := map[string]string{}
                            for _, a := range margs {
                                if eq := strings.IndexByte(a, '='); eq > 0 {
                                    kv[strings.ToLower(strings.TrimSpace(a[:eq]))] = strings.TrimSpace(a[eq+1:])
                                }
                            }
                            getInt := func(s string) int {
                                n := 0
                                for i := 0; i < len(s); i++ {
                                    if s[i] >= '0' && s[i] <= '9' { n = n*10 + int(s[i]-'0') } else { return 0 }
                                }
                                return n
                            }
                            capVal := 0
                            pol := ""
                            if v := kv["capacity"]; v != "" { capVal = getInt(v) } else if len(margs) > 0 { capVal = getInt(margs[0]) }
                            if v := kv["policy"]; v != "" { pol = v } else if len(margs) > 1 { pol = margs[1] }
                            // derive edge attrs
                            if capVal > 0 { op.Edge.Bounded = true }
                            if pol == "dropOldest" || pol == "dropNewest" { op.Edge.Delivery = "bestEffort" }
                            if pol == "block" { op.Edge.Delivery = "atLeastOnce" }
                            nb := struct{ Capacity int `json:"capacity"`; Policy string `json:"policy,omitempty"` }{Capacity: capVal, Policy: pol}
                            norm.Buffer = &nb
                        }
                        if at.Name == "merge.Stable" {
                            norm.Stable = true
                        }
                        if at.Name == "merge.Sort" {
                            var field, order string
                            if len(margs) > 0 { field = margs[0] }
                            if len(margs) > 1 { order = margs[1] } else { order = "asc" }
                            norm.Sort = append(norm.Sort, struct{
                                Field string `json:"field"`
                                Order string `json:"order"`
                            }{Field: field, Order: order})
                        }
                        if at.Name == "merge.Key" {
                            if len(margs) > 0 { norm.Key = margs[0] }
                        }
                        if at.Name == "merge.PartitionBy" {
                            if len(margs) > 0 { norm.PartitionBy = margs[0] }
                        }
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
                        if at.Name == "merge.Dedup" {
                            if len(margs) > 0 { norm.Dedup = margs[0] } else { norm.Dedup = "" }
                        }
                    }
                    if (at.Name == "edge.MultiPath" || at.Name == "MultiPath") && st.Name == "Collect" {
                        var margs []string
                        for _, aa := range at.Args { margs = append(margs, aa.Text) }
                        // collect inputs by scanning edges in this pipeline body
                        var inputs []string
                        for _, s2 := range pd.Stmts {
                            if e, ok := s2.(*ast.EdgeStmt); ok && e.To == st.Name {
                                inputs = append(inputs, e.From)
                            }
                        }
                        sort.Strings(inputs)
                        op.MultiPath = &pipeMultiPath{Args: margs, Inputs: inputs}
                    }
                }
                op.Attrs = rawAttrs
                if norm.Buffer != nil || norm.Stable || len(norm.Sort) > 0 ||
                    norm.Key != "" || norm.PartitionBy != "" || norm.TimeoutMs != 0 ||
                    norm.Window != 0 || norm.Watermark != nil || norm.Dedup != "" {
                    op.MergeNorm = &norm
                }
                steps = append(steps, op)
            }
            if e, ok := s.(*ast.EdgeStmt); ok {
                // Resolve fromId as the nearest occurrence at or before this edge stmt index
                // Resolve toId as the nearest occurrence at or after this edge stmt index
                esIdx := indexOfStmt(pd.Stmts, s)
                fromID := nearestOccBefore(occs[e.From], esIdx)
                toID := nearestOccAfter(occs[e.To], esIdx)
                edges = append(edges, pipeEdge{From: e.From, To: e.To, FromID: fromID, ToID: toID})
            }
        }
        // Assign MultiPath inputs per Collect instance based on resolved toId
        if len(edges) > 0 {
            // Build map of inputs per Collect ID
            inputsByID := map[int][]string{}
            for _, ed := range edges {
                if ed.To == "Collect" && ed.ToID > 0 {
                    inputsByID[ed.ToID] = append(inputsByID[ed.ToID], ed.From)
                }
            }
            // Update each Collect step's MultiPath Inputs
            for i := range steps {
                if steps[i].Name == "Collect" && steps[i].ID > 0 {
                    if ins, ok := inputsByID[steps[i].ID]; ok {
                        sort.Strings(ins)
                        if steps[i].MultiPath == nil { steps[i].MultiPath = &pipeMultiPath{} }
                        steps[i].MultiPath.Inputs = ins
                    }
                }
            }
        }
        // compute connectivity metadata
        var conn *pipeConn
        if len(edges) > 0 {
            // degree and adjacency
            deg := map[string]int{}
            adj := map[string][]string{}
            radj := map[string][]string{}
            for _, e := range edges {
                deg[e.From]++
                deg[e.To]++
                adj[e.From] = append(adj[e.From], e.To)
                radj[e.To] = append(radj[e.To], e.From)
            }
            // disconnected: any step name with deg==0
            var disc []string
            for _, st := range steps {
                if deg[st.Name] == 0 { disc = append(disc, st.Name) }
            }
            sort.Strings(disc)
            // reachability from ingress to egress
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
            // nodes that cannot reach egress: reverse search from egress
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
            // compile lists
            var notFromIngress []string
            var cannotToEgress []string
            for _, st := range steps {
                if !reach[st.Name] { notFromIngress = append(notFromIngress, st.Name) }
                if !reachToEgress[st.Name] { cannotToEgress = append(cannotToEgress, st.Name) }
            }
            sort.Strings(notFromIngress)
            sort.Strings(cannotToEgress)
            conn = &pipeConn{HasEdges: true, IngressToEgress: reach["egress"], Disconnected: disc, UnreachableFromIngress: notFromIngress, CannotReachEgress: cannotToEgress}
        }
        // deterministic ordering of edges for stability
        if len(edges) > 0 {
            sort.SliceStable(edges, func(i, j int) bool {
                if edges[i].From == edges[j].From { return edges[i].To < edges[j].To }
                return edges[i].From < edges[j].From
            })
        }
        entries = append(entries, pipelineEntry{Name: pd.Name, Steps: steps, Edges: edges, Conn: conn})
    }
    // if no pipelines parsed, synthesize a minimal entry to preserve defaults for tests/tools
    if len(entries) == 0 {
        entries = []pipelineEntry{{Name: "", Steps: []pipelineOp{{Name: "", Edge: &edgeAttrs{Bounded: false, Delivery: defaultDelivery}}}}}
    }
    // deterministic ordering by pipeline name
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
