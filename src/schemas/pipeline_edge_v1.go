package schemas

// PipelineEdgeV1 captures a declarative input edge for a pipeline step.
// Fields mirror the compiler-generated specs (FIFO/LIFO/Pipeline).
type PipelineEdgeV1 struct {
    Kind         string `json:"kind"` // edge.FIFO | edge.LIFO | edge.Pipeline
    MinCapacity  int    `json:"minCapacity,omitempty"`
    MaxCapacity  int    `json:"maxCapacity,omitempty"`
    Backpressure string `json:"backpressure,omitempty"`
    Type         string `json:"type,omitempty"`
    UpstreamName string `json:"upstreamName,omitempty"`
    // Derived semantics
    Bounded  bool   `json:"bounded,omitempty"`  // true when maxCapacity > 0; false indicates unbounded
    Delivery string `json:"delivery,omitempty"` // atLeastOnce (block) | bestEffort (drop)
}

