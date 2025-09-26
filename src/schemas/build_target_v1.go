package schemas

// BuildTarget represents a unit build target and its I/O.
type BuildTarget struct {
    Package string   `json:"package"`
    Unit    string   `json:"unit"`
    Inputs  []string `json:"inputs"`
    Outputs []string `json:"outputs"`
    Steps   []string `json:"steps"`
}

