package ast

// PipelineDecl represents a pipeline declaration and its node chain.
type PipelineDecl struct {
    Name            string
    Steps           []NodeCall
    Connectors      []string
    ErrorSteps      []NodeCall
    ErrorConnectors []string
    Pos             Position
    Comments        []Comment
}

func (PipelineDecl) isNode() {}

