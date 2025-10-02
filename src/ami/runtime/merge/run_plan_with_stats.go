package merge

import (
    "context"
    "time"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

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

