package graph

// Node represents a node in a pipeline graph.
// Kind examples: ingress, worker, egress, error.
type Node struct {
    ID    string
    Kind  string
    Label string
}

