package ast

// NodeCall is an invocation of a pipeline node with raw arguments.
type NodeCall struct {
    Name     string
    Args     []string
    Workers  []WorkerRef
    Pos      Position
    Comments []Comment
}

