package ir

// ErrorPipeline captures the error route declared inside a pipeline.
// It records the pipeline name and the ordered list of step names.
type ErrorPipeline struct {
    Pipeline string   `json:"pipeline"`
    Steps    []string `json:"steps"`
}

