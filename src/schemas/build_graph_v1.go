package schemas

// BuildGraph represents an optional DAG of build targets.
type BuildGraph struct {
    Nodes []string    `json:"nodes"`
    Edges [][2]string `json:"edges"`
}

