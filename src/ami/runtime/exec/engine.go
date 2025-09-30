package exec

import (
    "context"
    "fmt"
    "time"

    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    rmerge "github.com/sam-caldwell/ami/src/ami/runtime/merge"
    "github.com/sam-caldwell/ami/src/ami/runtime/scheduler"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// Engine configures and runs runtime tasks derived from IR modules.
type Engine struct {
    pool *scheduler.Pool
}

// NewEngineFromModule constructs a scheduler pool from an IR module's schedule and concurrency.
func NewEngineFromModule(m ir.Module) (*Engine, error) {
    pol, ok := scheduler.ParsePolicy(m.Schedule)
    if !ok || pol == "" { pol = scheduler.FIFO }
    workers := m.Concurrency
    if workers <= 0 { workers = 1 }
    p, err := scheduler.New(scheduler.Config{Workers: workers, Policy: pol})
    if err != nil { return nil, err }
    return &Engine{pool: p}, nil
}

// Close stops the scheduler pool, waiting for workers to exit.
func (e *Engine) Close() { if e.pool != nil { e.pool.Stop() } }

// RunMerge starts a merge task based on a MergePlan and returns an output channel.
// The caller should cancel ctx to stop the task; the returned channel is closed on shutdown.
func (e *Engine) RunMerge(ctx context.Context, plan ir.MergePlan, in <-chan ev.Event) (<-chan ev.Event, error) {
    if e == nil || e.pool == nil { return nil, fmt.Errorf("engine not initialized") }
    out := make(chan ev.Event, 1024)
    rp := toRuntimePlan(plan)
    if err := e.pool.Submit(rmerge.MergeTask("collect", rp, in, out)); err != nil { return nil, err }
    // Close 'out' when ctx is done (best-effort)
    go func(){ <-ctx.Done(); // allow loop to observe cancellation
        // Grace period to drain
        time.Sleep(5 * time.Millisecond)
        close(out)
    }()
    return out, nil
}

// RunMergeWithStats is like RunMerge but returns a Stats pointer populated at shutdown.
func (e *Engine) RunMergeWithStats(ctx context.Context, plan ir.MergePlan, in <-chan ev.Event) (<-chan ev.Event, *rmerge.Stats, error) {
    if e == nil || e.pool == nil { return nil, nil, fmt.Errorf("engine not initialized") }
    out := make(chan ev.Event, 1024)
    rp := toRuntimePlan(plan)
    var st rmerge.Stats
    // Inline runner to attach stats
    task := scheduler.Task{Source: "collect", Do: func(c context.Context){ rmerge.RunPlanWithStats(c, rp, in, out, &st) }}
    if err := e.pool.Submit(task); err != nil { return nil, nil, err }
    go func(){ <-ctx.Done(); time.Sleep(5 * time.Millisecond); close(out) }()
    return out, &st, nil
}

// Helper: convert IR MergePlan to runtime merge Plan.
func toRuntimePlan(p ir.MergePlan) rmerge.Plan {
    var rp rmerge.Plan
    rp.Stable = p.Stable
    for _, s := range p.Sort { rp.Sort = append(rp.Sort, rmerge.SortKey{Field: s.Field, Order: s.Order}) }
    rp.Key = p.Key
    rp.PartitionBy = p.PartitionBy
    rp.Buffer.Capacity = p.Buffer.Capacity
    rp.Buffer.Policy = p.Buffer.Policy
    rp.Window = p.Window
    rp.TimeoutMs = p.TimeoutMs
    rp.LatePolicy = p.LatePolicy
    if p.DedupField != "" { rp.Dedup.Field = p.DedupField }
    if p.Watermark != nil { rp.Watermark = &rmerge.Watermark{Field: p.Watermark.Field, LatenessMs: p.Watermark.LatenessMs} }
    return rp
}
