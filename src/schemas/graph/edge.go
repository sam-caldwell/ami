package graph

// Edge represents a directed connection from one node to another.
// Attrs may carry edge attributes (e.g., policy, capacity).
type Edge struct {
    From  string
    To    string
    Attrs map[string]any
}

