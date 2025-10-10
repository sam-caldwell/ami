package exec

import (
    "context"
    "time"
    "encoding/json"
    "os"
    "path/filepath"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    rmerge "github.com/sam-caldwell/ami/src/ami/runtime/merge"
    amitrigger "github.com/sam-caldwell/ami/src/ami/runtime/host/trigger"
    amiio "github.com/sam-caldwell/ami/src/ami/runtime/host/io"
    amitime "github.com/sam-caldwell/ami/src/ami/runtime/host/time"
    amigpu "github.com/sam-caldwell/ami/src/ami/runtime/host/gpu"
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
    type edgePolicy struct{ cap int; bp string; from string; to string }
    sendWithBP := func(next chan ev.Event, e ev.Event, st *rmerge.Stats, bp string, pol edgePolicy) {
        switch bp {
        case "dropNewest":
            select {
            case next <- e:
            default:
                if st != nil { st.Dropped++ }
            }
        case "dropOldest":
            select {
            case next <- e:
            default:
                // buffer full: evict one oldest then try once more
                select { case <-next: default: }
                select { case next <- e: default: if st != nil { st.Dropped++ } }
            }
        case "shuntNewest":
            // attempt send; if full, emit shunt advisory and drop from main stream
            select {
            case next <- e:
            default:
                if st != nil { st.Dropped++ }
                if opts.ErrorChan != nil {
                    select { case opts.ErrorChan <- errs.Error{Level: "warn", Code: "W_SHUNTED_NEWEST", Message: "shunted (newest)", Data: map[string]any{"from": pol.from, "to": pol.to}}: default: }
                }
            }
        case "shuntOldest":
            // attempt send; if full, evict one oldest and attempt once, else shunt
            select {
            case next <- e:
            default:
                // evict oldest
                select { case <-next: default: }
                select {
                case next <- e:
                default:
                    if st != nil { st.Dropped++ }
                    if opts.ErrorChan != nil {
                        select { case opts.ErrorChan <- errs.Error{Level: "warn", Code: "W_SHUNTED_OLDEST", Message: "shunted (oldest)", Data: map[string]any{"from": pol.from, "to": pol.to}}: default: }
                    }
                }
            }
        default: // block
            next <- e
        }
    }
    runTransform := func(idx int, name string, prev <-chan ev.Event, pol edgePolicy) <-chan ev.Event {
        buf := pol.cap
        if buf <= 0 { buf = 1024 }
        next := make(chan ev.Event, buf)
        var st rmerge.Stats
        go func(){
            for e := range prev {
                st.Enqueued++
                // Apply simple filter/transform stubs
                if keep := applyFilter(filterExpr, e); !keep { st.Dropped++; continue }
                e = applyTransform(transformExpr, e)
                sendWithBP(next, e, &st, pol.bp, pol); st.Emitted++
            }
            close(next)
            forwardStats(StageInfo{Name: name, Kind: "transform", Index: idx}, st)
        }()
        return next
    }
    runIngress := func(prev <-chan ev.Event, pol edgePolicy) <-chan ev.Event {
        buf := pol.cap
        if buf <= 0 { buf = 1024 }
        next := make(chan ev.Event, buf)
        var st rmerge.Stats
        go func(){ for e := range prev { st.Enqueued++; sendWithBP(next, e, &st, pol.bp, pol); st.Emitted++ }; close(next); forwardStats(StageInfo{Name:"ingress", Kind:"ingress", Index:0}, st) }()
        return next
    }
    // runGpuDispatch is a minimal GPU transform stage placeholder that integrates with
    // scheduling and honors sandbox policy. It performs backend preflight checks and
    // then passes events through unchanged to keep determinism in tests/environments
    // without a real GPU. Deterministic stats are emitted upon completion.
    runGpuDispatch := func(idx int, prev <-chan ev.Event, pol edgePolicy) (<-chan ev.Event, error) {
        if err := sandboxCheck(opts.Sandbox, "device"); err != nil { return nil, err }
        // Preflight probes: call availability helpers (results are advisory here)
        _ = amigpu.MetalAvailable()
        _ = amigpu.CudaAvailable()
        _ = amigpu.OpenCLAvailable()
        buf := pol.cap
        if buf <= 0 { buf = 1024 }
        next := make(chan ev.Event, buf)
        var st rmerge.Stats
        go func(){
            for e := range prev {
                st.Enqueued++
                // In this bring-up stage, dispatch is a no-op to preserve determinism
                // across hosts; the payload is forwarded unmodified.
                sendWithBP(next, e, &st, pol.bp, pol)
                st.Emitted++
            }
            close(next)
            forwardStats(StageInfo{Name: "GpuDispatch", Kind: "transform", Index: idx}, st)
        }()
        return next, nil
    }
    runEgress := func(prev <-chan ev.Event, pol edgePolicy) <-chan ev.Event {
        buf := pol.cap
        if buf <= 0 { buf = 1024 }
        next := make(chan ev.Event, buf)
        var st rmerge.Stats
        go func(){ for e := range prev { st.Enqueued++; sendWithBP(next, e, &st, pol.bp, pol); st.Emitted++ }; close(next); forwardStats(StageInfo{Name:"egress", Kind:"egress", Index:0}, st); close(statsOut) }()
        return next
    }
    // worker wrapper is inlined where needed to honor per-edge backpressure
    // Edges-based attempt
    if m.Package != "" {
        // Load transform worker names (if available) to bind them along the path deterministically.
        tnames, _ := loadTransformWorkers(".", m.Package, pipeline)
        // Load full edges index for policies
        var eidx edgesIndex
        if bnodes, err := BuildLinearPathFromEdges(".", m.Package, pipeline); err == nil && len(bnodes) > 0 {
            // read the same file used by BuildLinearPathFromEdges
            // we cannot get the path from it, so reconstruct filepath
        }
        // helper resolver: capacity/backpressure for pair
        resolveEP := func(from, to string) edgePolicy {
            // default
            pol := edgePolicy{cap: 1024, bp: "block", from: from, to: to}
            path := filepath.Join(".", "build", "debug", "asm", m.Package, "edges.json")
            if b, err := os.ReadFile(path); err == nil {
                if err := json.Unmarshal(b, &eidx); err == nil {
                    for _, ed := range eidx.Edges {
                        if ed.Pipeline != pipeline { continue }
                        if ed.From == from && ed.To == to {
                            if ed.MaxCapacity > 0 { pol.cap = ed.MaxCapacity }
                            // prefer explicit backpressure; fallback to delivery mapping
                            if ed.Backpressure != "" { pol.bp = ed.Backpressure } else {
                                switch ed.Delivery {
                                case "bestEffort": pol.bp = "dropNewest"
                                case "atLeastOnce": pol.bp = "block"
                                case "shuntNewest": pol.bp = "dropNewest"
                                case "shuntOldest": pol.bp = "dropOldest"
                                default:
                                }
                            }
                            break
                        }
                    }
                }
            }
            return pol
        }
        tIdxForNames := 0
        if nodes, err := BuildLinearPathFromEdges(".", m.Package, pipeline); err == nil && len(nodes) > 1 {
            cur := out
            // Optional Timer source when specified by opts or edges contain a Timer node
            hasTimer := false
            for _, n := range nodes { if n == "Timer" { hasTimer = true; break } }
            // Use timer when explicitly requested, or when present in edges and source=auto
            if hasTimer {
                if err := sandboxCheck(opts.Sandbox, "device"); err != nil { return nil, nil, err }
                // ingress -> Timer edge policy
                pol := resolveEP("ingress", "Timer")
                buf := pol.cap; if buf <= 0 { buf = 1024 }
                ch := make(chan ev.Event, buf)
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
                            e := ev.Event{Payload: map[string]any{"i": i, "ts": toStdTime(tm.Value)}}
                            sendWithBP(ch, e, &st, pol.bp, pol)
                            st.Emitted++
                            i++
                        }
                    }
                    close(ch); forwardStats(StageInfo{Name:"Timer", Kind:"ingress", Index:0}, st)
                }()
                cur = ch
            } else {
                pol := resolveEP("ingress", nodes[1])
                cur = runIngress(cur, pol)
            }
            tIdx := 0; cIdx := 0
            for i, name := range nodes {
                switch name {
                case "ingress":
                    // already wrapped
                case "egress":
                    // handled after loop
                case "Collect":
                    var mp *ir.MergePlan
                    for _, p := range m.Pipelines { if p.Name == pipeline { for _, c := range p.Collect { if c.Merge != nil { mp = c.Merge; break } } } }
                    if mp != nil {
                        // edge policy from previous node to Collect
                        prevName := nodes[i-1]
                        pol := resolveEP(prevName, "Collect")
                        buf := pol.cap; if buf <= 0 { buf = 1024 }
                        ch := make(chan ev.Event, buf)
                        go func(prev <-chan ev.Event, next chan ev.Event, pol edgePolicy){ for e := range prev { sendWithBP(next, e, nil, pol.bp, pol) }; close(next) }(cur, ch, pol)
                        oc, s, err := e.runMergeStageWithStats(ctx, *mp, ch)
                        if err != nil { return nil, nil, err }
                        go func(idx int, sp *rmerge.Stats){ <-ctx.Done(); forwardStats(StageInfo{Name:"Collect", Kind:"collect", Index:idx}, *sp) }(cIdx, s)
                        cIdx++
                        cur = oc
                    }
                default:
                    // If a worker name is available for this transform, run the worker; otherwise, identity transform+DSL stubs.
                    if name == "GpuDispatch" {
                        pol := resolveEP(name, nodes[i+1])
                        ch, err := runGpuDispatch(tIdx, cur, pol)
                        if err != nil { return nil, nil, err }
                        cur = ch
                    } else if name == "Transform" && tIdxForNames < len(tnames) && tnames[tIdxForNames] != "" {
                        pol := resolveEP(name, nodes[i+1])
                        // Wrap worker stage to honor backpressure on its output
                        next := func(prev <-chan ev.Event) <-chan ev.Event {
                            buf := pol.cap; if buf <= 0 { buf = 1024 }
                            outc := make(chan ev.Event, buf)
                            var st rmerge.Stats
                            wf := func(e ev.Event) (any, error) { return e, nil }
                            wname := tnames[tIdxForNames]
                            resolved := false
                            if effectiveInvoker != nil {
                                if f, ok := effectiveInvoker.Resolve(wname); ok && f != nil { wf = f; resolved = true }
                            }
                            if !resolved && opts.Workers != nil {
                                if f, ok := opts.Workers[wname]; ok && f != nil { wf = f; resolved = true }
                            }
                            go func(){
                                for e := range prev {
                                    st.Enqueued++
                                    out, err := wf(e)
                                    if err != nil {
                                        ee := errs.Error{Level: "error", Code: "E_WORKER", Message: err.Error(), Data: map[string]any{"worker": wname}}
                                        if opts.ErrorChan != nil { select { case opts.ErrorChan <- ee: default: } } else { ne := e; ne.Payload = ee; sendWithBP(outc, ne, &st, pol.bp, pol); st.Emitted++ }
                                        continue
                                    }
                                    switch v := out.(type) {
                                    case ev.Event:
                                        sendWithBP(outc, v, &st, pol.bp, pol)
                                    default:
                                        ne := e; ne.Payload = v; sendWithBP(outc, ne, &st, pol.bp, pol)
                                    }
                                    st.Emitted++
                                }
                                close(outc)
                                forwardStats(StageInfo{Name: wname, Kind: "transform", Index: tIdx}, st)
                            }()
                            return outc
                        }
                        cur = next(cur)
                        tIdxForNames++
                    } else {
                        pol := resolveEP(name, nodes[i+1])
                        cur = runTransform(tIdx, name, cur, pol)
                    }
                    tIdx++
                }
            }
            // wrap egress for stats
            // resolve policy for last hop to egress
            if len(nodes) >= 2 {
                pol := resolveEP(nodes[len(nodes)-2], "egress")
                cur = runEgress(cur, pol)
            } else {
                cur = runEgress(cur, edgePolicy{cap:1024, bp:"block"})
            }
            return cur, statsOut, nil
        }
    }
    // Fallback: IR collect order with transform stubs as identity
    var cur <-chan ev.Event
    // Rely on IR-defined ingress; do not synthesize sources here
    cur = runIngress(out, edgePolicy{cap: 1024, bp: "block"})
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
        cur = runEgress(cur, edgePolicy{cap: 1024, bp: "block"})
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
