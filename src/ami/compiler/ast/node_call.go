package ast

// NodeCall is an invocation of a pipeline node.
// Args retains the original raw argument strings for backward compatibility.
// Attrs holds structured key/value attributes parsed from args (e.g., in=, worker=, minWorkers, ...).
type NodeCall struct {
    Name     string
    Args     []string
    Attrs    map[string]string
    InlineWorker *FuncLit
    Workers  []WorkerRef
    Pos      Position
    Comments []Comment
}
