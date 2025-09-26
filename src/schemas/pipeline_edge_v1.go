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
    // MultiPath scaffold: when Kind == "edge.MultiPath", inputs and merge ops are provided below.
    MultiPath *MultiPathV1 `json:"multiPath,omitempty"`
}

// MultiPathV1 carries a list of input edges and optional merge operations.
type MultiPathV1 struct {
    Inputs []PipelineEdgeV1 `json:"inputs"`
    Merge  []MergeOpV1      `json:"merge,omitempty"`
}

// MergeOpV1 is a minimal representation of a merge operation.
type MergeOpV1 struct {
    Name string `json:"name"`
    Raw  string `json:"raw,omitempty"`
}
