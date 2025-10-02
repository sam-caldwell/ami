package driver

type importDetail struct {
    Path       string `json:"path"`
    Constraint string `json:"constraint,omitempty"`
}

