package exec

import (
    "context"
    "time"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    rmerge "github.com/sam-caldwell/ami/src/ami/runtime/merge"
    amitrigger "github.com/sam-caldwell/ami/src/ami/stdlib/trigger"
    amitime "github.com/sam-caldwell/ami/src/ami/stdlib/time"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// RunPipeline chains Collect merge steps sequentially for the named pipeline.
// Returns the final output channel; caller cancels ctx to stop.
func (e *Engine) RunPipeline(ctx context.Context, m ir.Module, pipeline string, in <-chan ev.Event) (<-chan ev.Event, error) {
    var out <-chan ev.Event = in
    // Attempt edges-based order for higher fidelity; fallback to IR collect order.
    if m.Package != "" {
        if nodes, err := BuildLinearPathFromEdges(".", m.Package, pipeline); err == nil && len(nodes) > 0 {
            // Walk nodes; treat Collect specially; other unknowns: Transform pass-through
            cur := out
            for _, name := range nodes {
                switch name {
                case "ingress":
                    // already handled by 'in'
                case "egress":
                    // terminal; no additional stage
                case "Collect":
                    // choose the first Collect plan in IR for now (simple mapping)
                    var mp *ir.MergePlan
                    for _, p := range m.Pipelines { if p.Name == pipeline { for _, c := range p.Collect { if c.Merge != nil { mp = c.Merge; break } } } }
                    if mp != nil {
                        ch := make(chan ev.Event, 1024)
                        go func(prev <-chan ev.Event, next chan<- ev.Event){ for e := range prev { next <- e }; close(next) }(cur, ch)
                        oc, err := e.RunMerge(ctx, *mp, ch); if err != nil { return nil, err }
                        cur = oc
                    }
                default:
                    // Transform: pass-through
                    ch := make(chan ev.Event, 1024)
                    go func(prev <-chan ev.Event, next chan<- ev.Event){ for e := range prev { next <- e }; close(next) }(cur, ch)
                    cur = ch
                }
            }
            return cur, nil
        }
    }
    // Fallback: IR collect chain order
    var cur <-chan ev.Event = out
    for _, p := range m.Pipelines {
        if p.Name != pipeline { continue }
        for _, c := range p.Collect {
            if c.Merge == nil { continue }
            ch := make(chan ev.Event, 1024)
            go func(prev <-chan ev.Event, next chan<- ev.Event){ for e := range prev { next <- e }; close(next) }(cur, ch)
            oc, err := e.RunMerge(ctx, *c.Merge, ch)
            if err != nil { return nil, err }
            cur = oc
        }
        return cur, nil
    }
    return out, nil
}

// RunPipelineWithStats chains stages and emits stage-level stats via emit callback.
// filterExpr/transformExpr are simple DSL stubs applied at Transform stages.
func (e *Engine) RunPipelineWithStats(ctx context.Context, m ir.Module, pipeline string, in <-chan ev.Event, emit func(StageInfo, rmerge.Stats), filterExpr, transformExpr string, opts ExecOptions) (<-chan ev.Event, <-chan StageStats, error) {
    statsOut := make(chan StageStats, 16)
    forwardStats := func(info StageInfo, st rmerge.Stats){ if emit != nil { emit(info, st) }; statsOut <- StageStats{Stage: info, Stats: st} }
    var out <-chan ev.Event = in
    // helper: transform pass-through with counters and DSL filtering/transforming
    runTransform := func(idx int, name string, prev <-chan ev.Event) <-chan ev.Event {
        next := make(chan ev.Event, 1024)
        var st rmerge.Stats
        go func(){
            for e := range prev {
                st.Enqueued++
                // Apply simple filter/transform stubs
                if keep := applyFilter(filterExpr, e); !keep { st.Dropped++; continue }
                e = applyTransform(transformExpr, e)
                next <- e; st.Emitted++
            }
            close(next)
            forwardStats(StageInfo{Name: name, Kind: "transform", Index: idx}, st)
        }()
        return next
    }
    runIngress := func(prev <-chan ev.Event) <-chan ev.Event {
        next := make(chan ev.Event, 1024)
        var st rmerge.Stats
        go func(){ for e := range prev { st.Enqueued++; next <- e; st.Emitted++ }; close(next); forwardStats(StageInfo{Name:"ingress", Kind:"ingress", Index:0}, st) }()
        return next
    }
    runEgress := func(prev <-chan ev.Event) <-chan ev.Event {
        next := make(chan ev.Event, 1024)
        var st rmerge.Stats
        go func(){ for e := range prev { st.Enqueued++; next <- e; st.Emitted++ }; close(next); forwardStats(StageInfo{Name:"egress", Kind:"egress", Index:0}, st); close(statsOut) }()
        return next
    }
    // Edges-based attempt
    if m.Package != "" {
        if nodes, err := BuildLinearPathFromEdges(".", m.Package, pipeline); err == nil && len(nodes) > 0 {
            cur := out
            // Optional Timer source when specified by opts or edges contain a Timer node
            hasTimer := false
            for _, n := range nodes { if n == "Timer" { hasTimer = true; break } }
            // Use timer when explicitly requested, or when present in edges and source=auto
            if hasTimer {
                if err := sandboxCheck(opts.Sandbox, "device"); err != nil { return nil, nil, err }
                ch := make(chan ev.Event, 1024)
                var st rmerge.Stats
                go func(){
                    // Use AMI trigger.Timer to emit time events and adapt to schemas.Event
                    tCh, stop := amitrigger.Timer(amitime.Duration(opts.TimerInterval))
                    defer stop()
                    i := 0
                    for {
                        if opts.TimerCount > 0 && i >= opts.TimerCount { break }
                        select {
                        case <-ctx.Done():
                            break
                        case tm := <-tCh:
                            st.Enqueued++
                            ch <- ev.Event{Payload: map[string]any{"i": i, "ts": toStdTime(tm.Value)}}
                            st.Emitted++
                            i++
                        }
                    }
                    close(ch); forwardStats(StageInfo{Name:"Timer", Kind:"ingress", Index:0}, st)
                }()
                cur = ch
            } else {
                cur = runIngress(cur)
            }
            tIdx := 0; cIdx := 0
            for _, name := range nodes {
                switch name {
                case "ingress":
                    // already wrapped
                case "egress":
                    // handled after loop
                case "Collect":
                    var mp *ir.MergePlan
                    for _, p := range m.Pipelines { if p.Name == pipeline { for _, c := range p.Collect { if c.Merge != nil { mp = c.Merge; break } } } }
                    if mp != nil {
                        ch := make(chan ev.Event, 1024)
                        go func(prev <-chan ev.Event, next chan<- ev.Event){ for e := range prev { next <- e }; close(next) }(cur, ch)
                        oc, s, err := e.runMergeStageWithStats(ctx, *mp, ch)
                        if err != nil { return nil, nil, err }
                        go func(idx int, sp *rmerge.Stats){ <-ctx.Done(); forwardStats(StageInfo{Name:"Collect", Kind:"collect", Index:idx}, *sp) }(cIdx, s)
                        cIdx++
                        cur = oc
                    }
                default:
                    cur = runTransform(tIdx, name, cur)
                    tIdx++
                }
            }
            // wrap egress for stats
            cur = runEgress(cur)
            return cur, statsOut, nil
        }
    }
    // Fallback: IR collect order with transform stubs as identity
    var cur <-chan ev.Event
    // Rely on IR-defined ingress; do not synthesize sources here
    cur = runIngress(out)
    cIdx := 0
    for _, p := range m.Pipelines {
        if p.Name != pipeline { continue }
        for _, c := range p.Collect {
            if c.Merge == nil { continue }
            ch := make(chan ev.Event, 1024)
            go func(prev <-chan ev.Event, next chan<- ev.Event){ for e := range prev { next <- e }; close(next) }(cur, ch)
            oc, s, err := e.runMergeStageWithStats(ctx, *c.Merge, ch)
            if err != nil { return nil, nil, err }
            go func(idx int, sp *rmerge.Stats){ <-ctx.Done(); forwardStats(StageInfo{Name:"Collect", Kind:"collect", Index:idx}, *sp) }(cIdx, s)
            cIdx++
            cur = oc
        }
        cur = runEgress(cur)
        return cur, statsOut, nil
    }
    // No nodes; still emit egress stats on cancellation
    _ = time.Now() // keep import used
    go func(){ <-ctx.Done(); forwardStats(StageInfo{Name:"egress", Kind:"egress", Index:0}, rmerge.Stats{}); close(statsOut) }()
    return out, statsOut, nil
}

func (e *Engine) runMergeStageWithStats(ctx context.Context, plan ir.MergePlan, in <-chan ev.Event) (<-chan ev.Event, *rmerge.Stats, error) {
    oc, st, err := e.RunMergeWithStats(ctx, plan, in)
    if err != nil { return nil, nil, err }
    _ = time.Now() // placeholder to avoid unused imports if needed
    return oc, st, nil
}

// toStdTime converts amitime.Time to stdlib time.Time for payload compatibility.
func toStdTime(t amitime.Time) time.Time {
    sec := t.Unix()
    nsec := t.UnixNano() - sec*1_000_000_000
    return time.Unix(sec, nsec).UTC()
}

// DSL stubs
func applyFilter(expr string, e ev.Event) bool {
    switch expr {
    case "", "none":
        return true
    case "drop_even":
        // drops events where payload["i"] is even
        if m, ok := e.Payload.(map[string]any); ok {
            if v, ok := m["i"].(int); ok { return v%2 != 0 }
            if f, ok := m["i"].(float64); ok { return int(f)%2 != 0 }
        }
        return true
    default:
        return true
    }
}

func applyTransform(expr string, e ev.Event) ev.Event {
    switch expr {
    case "", "none":
        return e
    default:
        // add_field:name sets payload[name]=true
        if len(expr) > 10 && expr[:10] == "add_field:" {
            key := expr[10:]
            if m, ok := e.Payload.(map[string]any); ok && key != "" { m[key] = true; e.Payload = m }
        }
        return e
    }
}
