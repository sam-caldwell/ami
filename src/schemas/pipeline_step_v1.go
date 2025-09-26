package schemas

// PipelineStepV1 is a step within a pipeline.
type PipelineStepV1 struct {
    Node    string             `json:"node"`
    Attrs   map[string]string  `json:"attrs,omitempty"`
    Workers []PipelineWorkerV1 `json:"workers,omitempty"`
    InEdge  *PipelineEdgeV1    `json:"inEdge,omitempty"`
}
