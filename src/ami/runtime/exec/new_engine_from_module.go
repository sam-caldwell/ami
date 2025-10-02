package exec

import (
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/runtime/scheduler"
)

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

