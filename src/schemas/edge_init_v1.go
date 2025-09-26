package schemas

// EdgeInitV1 describes an input edge initialization for a pipeline step.
type EdgeInitV1 struct {
    Unit         string `json:"unit"`
    Pipeline     string `json:"pipeline"`
    Segment      string `json:"segment"` // normal|error
    Step         int    `json:"step"`
    Node         string `json:"node"`
    Label        string `json:"label"`
    Kind         string `json:"kind"`
    MinCapacity  int    `json:"minCapacity,omitempty"`
    MaxCapacity  int    `json:"maxCapacity,omitempty"`
    Backpressure string `json:"backpressure,omitempty"`
    Type         string `json:"type,omitempty"`
    UpstreamName string `json:"upstreamName,omitempty"`
    // Derived semantics
    Bounded  bool   `json:"bounded,omitempty"`  // true when maxCapacity > 0; false indicates unbounded
    Delivery string `json:"delivery,omitempty"` // atLeastOnce (block) | bestEffort (drop)
    // When Kind == "edge.MultiPath", include normalized inputs and merge ops.
    // Mirrors PipelineEdgeV1.MultiPath for debug parity with pipelines.v1.
    MultiPath *MultiPathV1 `json:"multiPath,omitempty"`
}
