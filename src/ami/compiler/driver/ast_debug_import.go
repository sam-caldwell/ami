package driver

type astImport struct {
    Path       string `json:"path"`
    Constraint string `json:"constraint,omitempty"`
    Pos        *dbgPos `json:"pos,omitempty"`
    Kind       string  `json:"kind"`
}

