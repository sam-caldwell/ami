package ast

// Directive captures a top-level `#pragma` directive and its payload.
type Directive struct {
    Name     string
    Payload  string
    Pos      Position
    Comments []Comment
}

func (Directive) isNode() {}

