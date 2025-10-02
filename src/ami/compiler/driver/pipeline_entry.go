package driver

type pipelineEntry struct {
    Name  string       `json:"name"`
    Steps []pipelineOp `json:"steps"`
    Edges []pipeEdge   `json:"edges,omitempty"`
    Conn  *pipeConn    `json:"connectivity,omitempty"`
}

