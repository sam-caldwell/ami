package schemas

// PipelineV1 represents a single pipeline with steps.
type PipelineV1 struct {
    Name       string           `json:"name"`
    Steps      []PipelineStepV1 `json:"steps"`
    ErrorSteps []PipelineStepV1 `json:"errorSteps,omitempty"`
}

