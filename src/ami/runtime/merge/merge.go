package merge

import (
    "errors"
)

var ErrFull = errors.New("merge buffer full")

type Config struct {
    BufferCapacity     int
    BufferBackpressure string // block|dropOldest|dropNewest
    Stable             bool
}

// Orchestrator is a minimal deterministic buffer orchestrator.
type Orchestrator struct {
    cfg   Config
    queue []any
}

func New(cfg Config) *Orchestrator {
    return &Orchestrator{cfg: cfg}
}

// AddInput is a no-op in this scaffold; inputs are tracked at codegen level.
func (o *Orchestrator) AddInput(_ int) {}

// Push enqueues a value honoring backpressure policy deterministically.
func (o *Orchestrator) Push(v any) error {
    cap := o.cfg.BufferCapacity
    if cap <= 0 {
        // unbounded; accept
        o.queue = append(o.queue, v)
        return nil
    }
    if len(o.queue) < cap {
        o.queue = append(o.queue, v)
        return nil
    }
    switch o.cfg.BufferBackpressure {
    case "dropOldest":
        // evict oldest and append
        copy(o.queue[0:], o.queue[1:])
        o.queue = o.queue[:cap-1]
        o.queue = append(o.queue, v)
        return nil
    case "dropNewest":
        // drop incoming, keep existing
        return nil
    default:
        return ErrFull
    }
}

// Pop returns the oldest enqueued value, FIFO.
func (o *Orchestrator) Pop() (any, bool) {
    if len(o.queue) == 0 { return nil, false }
    v := o.queue[0]
    o.queue = o.queue[1:]
    return v, true
}

