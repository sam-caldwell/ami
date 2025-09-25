package schemas

import "errors"

// PipelinesV1 is a debug schema that captures lowered pipeline structure
// for a compilation unit, including referenced workers and generic payloads.
// It is intended for tooling and golden tests and is emitted under
// build/debug/ir/<package>/<unit>.pipelines.json when verbose build is enabled.
type PipelinesV1 struct {
    Schema    string          `json:"schema"`
    Timestamp string          `json:"timestamp"`
    Package   string          `json:"package"`
    File      string          `json:"file"`
    Pipelines []PipelineV1    `json:"pipelines"`
}

type PipelineV1 struct {
    Name        string           `json:"name"`
    Steps       []PipelineStepV1 `json:"steps"`
    ErrorSteps  []PipelineStepV1 `json:"errorSteps,omitempty"`
}

type PipelineStepV1 struct {
    Node    string            `json:"node"`
    Workers []PipelineWorkerV1 `json:"workers,omitempty"`
    InEdge  *PipelineEdgeV1    `json:"inEdge,omitempty"`
}

type PipelineWorkerV1 struct {
    Name        string `json:"name"`
    Kind        string `json:"kind"` // function|factory
    HasContext  bool   `json:"hasContext"`
    HasState    bool   `json:"hasState"`
    // Generic payload type information captured from the signature shapes.
    Input       string `json:"input,omitempty"`        // Event<T> -> T
    OutputKind  string `json:"outputKind,omitempty"`   // Event|Events|Error|Drop|Ack
    Output      string `json:"output,omitempty"`       // Event<U>/Events<U>/Error<E> -> U/E
}

func (p *PipelinesV1) Validate() error {
    if p == nil { return errors.New("nil pipelines") }
    if p.Schema == "" { p.Schema = "pipelines.v1" }
    if p.Schema != "pipelines.v1" { return errors.New("invalid schema") }
    return nil
}

// PipelineEdgeV1 captures a declarative input edge for a pipeline step.
// Fields mirror the compiler-generated specs (FIFO/LIFO/Pipeline).
type PipelineEdgeV1 struct {
    Kind         string `json:"kind"` // edge.FIFO | edge.LIFO | edge.Pipeline
    MinCapacity  int    `json:"minCapacity,omitempty"`
    MaxCapacity  int    `json:"maxCapacity,omitempty"`
    Backpressure string `json:"backpressure,omitempty"`
    Type         string `json:"type,omitempty"`
    UpstreamName string `json:"upstreamName,omitempty"`
}
