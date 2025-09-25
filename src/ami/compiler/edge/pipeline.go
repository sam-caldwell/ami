package edge

import "errors"

// Pipeline is a declarative, compiler-generated bridge between pipelines.
// It identifies the upstream pipeline by its ingress name and declares the
// capacity/backpressure policy of the bridging queue.
type Pipeline struct {
    UpstreamName string
    MinCapacity  int
    MaxCapacity  int
    Backpressure BackpressurePolicy
    TypeName     string // payload type name for analysis/codegen
}

func (p Pipeline) Kind() string { return "edge.Pipeline" }

func (p Pipeline) Validate() error {
    if p.UpstreamName == "" { return errors.New("upstream pipeline name required") }
    if p.MinCapacity < 0 { return errors.New("minCapacity must be >= 0") }
    if p.MaxCapacity < p.MinCapacity { return errors.New("maxCapacity must be >= minCapacity") }
    switch p.Backpressure {
    case "", BackpressureBlock, BackpressureDrop:
    default:
        return errors.New("invalid backpressure policy")
    }
    return nil
}

