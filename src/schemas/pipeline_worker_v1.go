package schemas

// PipelineWorkerV1 captures a worker referenced by a step.
type PipelineWorkerV1 struct {
    Name       string `json:"name"`
    Kind       string `json:"kind"` // function|factory
    Origin     string `json:"origin,omitempty"` // reference|literal
    HasContext bool   `json:"hasContext"`
    HasState   bool   `json:"hasState"`
    // Generic payload type information captured from the signature shapes.
    Input      string `json:"input,omitempty"`      // Event<T> -> T
    OutputKind string `json:"outputKind,omitempty"` // Event|Events|Error
    Output     string `json:"output,omitempty"`     // Event<U>/Events<U>/Error<E> -> U/E
}
