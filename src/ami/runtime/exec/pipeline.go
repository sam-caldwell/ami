package exec

import (
    "context"
    "time"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    rmerge "github.com/sam-caldwell/ami/src/ami/runtime/merge"
    amitrigger "github.com/sam-caldwell/ami/src/ami/runtime/host/trigger"
    amiio "github.com/sam-caldwell/ami/src/ami/runtime/host/io"
    amitime "github.com/sam-caldwell/ami/src/ami/runtime/host/time"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
    errs "github.com/sam-caldwell/ami/src/schemas/errors"
)

// RunPipeline chains Collect merge steps sequentially for the named pipeline.
// Returns the final output channel; caller cancels ctx to stop.
func (e *Engine) RunPipeline(ctx context.Context, m ir.Module, pipeline string, in <-chan ev.Event) (<-chan ev.Event, error) {
    var out <-chan ev.Event = in
    // Attempt edges-based order for higher fidelity; fallback to IR collect order.
    if m.Package != "" {
        if nodes, err := BuildLinearPathFromEdges(".", m.Package, pipeline); err == nil && len(nodes) > 1 {
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
    forwardStats := func(info StageInfo, st rmerge.Stats){
        if emit != nil { emit(info, st) }
        defer func(){ _ = recover() }()
        statsOut <- StageStats{Stage: info, Stats: st}
    }
    var out <-chan ev.Event = in
    // Align stdlib io capabilities with sandbox policy for the duration of this run.
    prev := amiio.GetPolicy()
    amiio.SetPolicy(amiio.Policy{AllowFS: opts.Sandbox.AllowFS, AllowNet: opts.Sandbox.AllowNet, AllowDevice: opts.Sandbox.AllowDevice})
    defer amiio.SetPolicy(prev)
    // Prefer a dynamic invoker from manifest when not provided.
    effectiveInvoker := opts.Invoker
    if effectiveInvoker == nil && m.Package != "" {
        if lib := loadWorkersLibFromManifest(".", m.Package); lib != "" {
            if inv := NewDLSOInvoker(lib, "ami_worker_"); inv != nil {
                effectiveInvoker = inv
            }
        }
    }
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
    // Worker-based transform stage. Looks up worker by name via dynamic invoker first,
    // then falls back to opts.Workers registry if not found.
    runWorker := func(idx int, workerName string, prev <-chan ev.Event) <-chan ev.Event {
        next := make(chan ev.Event, 1024)
        var st rmerge.Stats
        wf := func(e ev.Event) (any, error) { return e, nil }
        // Prefer dynamic invoker when available; remember if resolution succeeded
        resolved := false
        if effectiveInvoker != nil {
            if f, ok := effectiveInvoker.Resolve(workerName); ok && f != nil {
                wf = f
                resolved = true
            }
        }
        // Fallback to in-process registry if invoker not found the worker
        if !resolved && opts.Workers != nil {
            if f, ok := opts.Workers[workerName]; ok && f != nil {
                wf = f
                resolved = true
            }
        }
        go func(){
            for e := range prev {
                st.Enqueued++
                out, err := wf(e)
                if err != nil {
                    ee := errs.Error{Level: "error", Code: "E_WORKER", Message: err.Error(), Data: map[string]any{"worker": workerName}}
                    if opts.ErrorChan != nil {
                        // Emit via error channel and do not inject into main event stream.
                        select { case opts.ErrorChan <- ee: default: /* drop if not drained */ }
                    } else {
                        ne := e
                        ne.Payload = ee
                        next <- ne
                        st.Emitted++
                    }
                    continue
                }
                switch v := out.(type) {
                case ev.Event:
                    next <- v
                default:
                    ne := e
                    ne.Payload = v
                    next <- ne
                }
                st.Emitted++
            }
            close(next)
            forwardStats(StageInfo{Name: workerName, Kind: "transform", Index: idx}, st)
        }()
        return next
    }
    // Edges-based attempt
    if m.Package != "" {
        // Load transform worker names (if available) to bind them along the path deterministically.
        tnames, _ := loadTransformWorkers(".", m.Package, pipeline)
        tIdxForNames := 0
        if nodes, err := BuildLinearPathFromEdges(".", m.Package, pipeline); err == nil && len(nodes) > 1 {
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
                    // If a worker name is available for this transform, run the worker; otherwise, identity transform+DSL stubs.
                    if name == "Transform" && tIdxForNames < len(tnames) && tnames[tIdxForNames] != "" {
                        cur = runWorker(tIdx, tnames[tIdxForNames], cur)
                        tIdxForNames++
                    } else {
                        cur = runTransform(tIdx, name, cur)
                    }
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
