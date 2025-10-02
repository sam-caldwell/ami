package driver

type resolvedUnit struct {
    Package string   `json:"package"`
    File    string   `json:"file"`
    Imports []string `json:"imports,omitempty"`
    Source  string   `json:"source"`
}

