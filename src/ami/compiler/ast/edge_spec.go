package ast

// EdgeSpec captures an explicit edge with policy and configuration.
type EdgeSpec struct {
    From         string
    To           string
    Connector    string
    Backpressure string
    BufferSize   int
}

func (EdgeSpec) isNode() {}

