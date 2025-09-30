package merge

import (
    "context"
    "time"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// RunPlan executes a simple merge loop: it reads events from in, applies the
// merge Plan, and emits ordered events to out until ctx is done. It periodically
// applies TimeoutMs expiration. This is a cooperative, single-goroutine loop
// suitable for testing and as a scheduler hook.
func RunPlan(ctx context.Context, plan Plan, in <-chan ev.Event, out chan<- ev.Event) {
    // In streaming runloop, default to accepting late arrivals and letting watermark flush determine emission.
    if plan.Watermark != nil && plan.Watermark.LatenessMs > 0 && plan.LatePolicy == "" { plan.LatePolicy = "accept" }
    op := NewOperator(plan)
    tick := time.NewTicker(10 * time.Millisecond)
    defer tick.Stop()
    for {
        // First, enforce window and watermark flushes to preserve ordering under pressure.
        if flushed := op.FlushWindowExcess(); len(flushed) > 0 {
            for _, fe := range flushed { out <- fe }
        }
        if flushed := op.FlushByWatermark(time.Now()); len(flushed) > 0 {
            for _, fe := range flushed { out <- fe }
        }
        select {
        case <-ctx.Done():
            // Drain remaining items in order before exit.
            for { if e, ok := op.Pop(); ok { out <- e } else { break } }
            return
        case e, ok := <-in:
            if !ok {
                // input closed: drain and exit
                for { if x, ok := op.Pop(); ok { out <- x } else { break } }
                return
            }
            // Best-effort push; on backpressure, emit one and retry.
            if err := op.Push(e); err == ErrBackpressure { if x, ok := op.Pop(); ok { out <- x }; _ = op.Push(e) }
            // Continue loop to prefer batching more inputs before emitting via Pop.
            continue
        case now := <-tick.C:
            _ = op.ExpireStale(now)
            // also enforce window on tick
            if flushed := op.FlushWindowExcess(); len(flushed) > 0 { for _, fe := range flushed { out <- fe } }
            continue
        default:
            // Idle: emit next available item in sorted order.
            if e, ok := op.Pop(); ok { out <- e } else { time.Sleep(1 * time.Millisecond) }
        }
    }
}

// Stats captures operator counters for reporting.
type Stats struct{ Enqueued, Emitted, Dropped, Expired int64 }

// RunPlanWithStats is like RunPlan but updates stats at exit.
func RunPlanWithStats(ctx context.Context, plan Plan, in <-chan ev.Event, out chan<- ev.Event, stats *Stats) {
    if plan.Watermark != nil && plan.Watermark.LatenessMs > 0 && plan.LatePolicy == "" { plan.LatePolicy = "accept" }
    op := NewOperator(plan)
    tick := time.NewTicker(10 * time.Millisecond)
    defer tick.Stop()
    for {
        if flushed := op.FlushWindowExcess(); len(flushed) > 0 { for _, fe := range flushed { out <- fe } }
        if flushed := op.FlushByWatermark(time.Now()); len(flushed) > 0 { for _, fe := range flushed { out <- fe } }
        select {
        case <-ctx.Done():
            if stats != nil { enq, emit, drop, exp := op.Stats(); stats.Enqueued, stats.Emitted, stats.Dropped, stats.Expired = enq, emit, drop, exp }
            for { if x, ok := op.Pop(); ok { out <- x } else { break } }
            return
        case e, ok := <-in:
            if !ok {
                if stats != nil { enq, emit, drop, exp := op.Stats(); stats.Enqueued, stats.Emitted, stats.Dropped, stats.Expired = enq, emit, drop, exp }
                for { if x, ok := op.Pop(); ok { out <- x } else { break } }
                return
            }
            if err := op.Push(e); err == ErrBackpressure { if x, ok := op.Pop(); ok { out <- x }; _ = op.Push(e) }
            continue
        case now := <-tick.C:
            _ = op.ExpireStale(now)
            continue
        default:
            if e, ok := op.Pop(); ok { out <- e } else { time.Sleep(1 * time.Millisecond) }
        }
    }
}
