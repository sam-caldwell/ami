package ir

// Directive captures a compiler directive (pragma) for IR debug.
type Directive struct {
    Domain string            `json:"domain"`
    Key    string            `json:"key,omitempty"`
    Value  string            `json:"value,omitempty"`
    Args   []string          `json:"args,omitempty"`
    Params map[string]string `json:"params,omitempty"`
}

