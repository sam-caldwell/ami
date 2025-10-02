package driver

type pragmaDetail struct {
    Line   int               `json:"line"`
    Domain string            `json:"domain"`
    Key    string            `json:"key"`
    Value  string            `json:"value,omitempty"`
    Params map[string]string `json:"params,omitempty"`
}

