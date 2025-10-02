package scheduler

import (
    "context"
    "errors"
)

var ErrInvalidWorkers = errors.New("workers must be >= 1")

func New(cfg Config) (*Pool, error) {
    if cfg.Workers <= 0 { return nil, ErrInvalidWorkers }
    if cfg.Policy == "" { cfg.Policy = FIFO }
    ctx, cancel := context.WithCancel(context.Background())
    p := &Pool{cfg: cfg, ctx: ctx, cancel: cancel}
    switch cfg.Policy {
    case FIFO:
        // Default to a small buffered channel so Submit does not spuriously fail
        // when no explicit QueueCapacity is provided. An unbuffered channel with
        // non-blocking Submit would frequently return "queue full" under test.
        n := cfg.QueueCapacity
        if n <= 0 {
            n = 64
        }
        p.fifoCh = make(chan Task, n)
    case LIFO:
        p.lifo = make([]Task, 0)
    case FAIR:
        p.fairQ = make(map[string][]Task)
        p.fairOrder = make([]string, 0)
    case WORKSTEAL:
        p.wsQ = make([]chan Task, cfg.Workers)
        for i := 0; i < cfg.Workers; i++ {
            // small bounded channels to exercise stealing
            p.wsQ[i] = make(chan Task, 64)
        }
    }
    p.spawn()
    return p, nil
}

