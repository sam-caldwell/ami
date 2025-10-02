package driver

type astPragma struct {
    Domain string            `json:"domain,omitempty"`
    Key    string            `json:"key,omitempty"`
    Value  string            `json:"value,omitempty"`
    Args   []string          `json:"args,omitempty"`
    Params map[string]string `json:"params,omitempty"`
    Pos    *dbgPos           `json:"pos,omitempty"`
    Kind   string            `json:"kind"`
}

