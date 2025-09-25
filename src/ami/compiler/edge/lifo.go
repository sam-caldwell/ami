package edge

import "errors"

// LIFO is a declarative, compiler-generated last-in, first-out edge.
// Codegen emits a high-performance stack-like ring optimized per payload type.
type LIFO struct {
    MinCapacity int
    MaxCapacity int
    Backpressure BackpressurePolicy
    TypeName     string // payload type name for analysis/codegen
}

func (l LIFO) Kind() string { return "edge.LIFO" }

func (l LIFO) Validate() error {
    if l.MinCapacity < 0 { return errors.New("minCapacity must be >= 0") }
    if l.MaxCapacity < l.MinCapacity { return errors.New("maxCapacity must be >= minCapacity") }
    switch l.Backpressure {
    case "", BackpressureBlock, BackpressureDrop:
    default:
        return errors.New("invalid backpressure policy")
    }
    return nil
}

