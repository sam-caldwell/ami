package edge

import "errors"

// FIFO is a declarative, compiler-generated first-in, first-out edge.
// Codegen emits a high-performance MPMC ring buffer specialized to the payload
// type and capacity bounds.
type FIFO struct {
    MinCapacity int
    MaxCapacity int
    Backpressure BackpressurePolicy
    TypeName     string // payload type name for analysis/codegen
}

func (f FIFO) Kind() string { return "edge.FIFO" }

func (f FIFO) Validate() error {
    if f.MinCapacity < 0 { return errors.New("minCapacity must be >= 0") }
    if f.MaxCapacity < f.MinCapacity { return errors.New("maxCapacity must be >= minCapacity") }
    switch f.Backpressure {
    case "", BackpressureBlock, BackpressureDrop:
    default:
        return errors.New("invalid backpressure policy")
    }
    return nil
}

