package driver

type pipelineList struct {
    Schema      string          `json:"schema"`
    Package     string          `json:"package"`
    Unit        string          `json:"unit"`
    Concurrency *pipeConcurrency `json:"concurrency,omitempty"`
    Pipelines   []pipelineEntry `json:"pipelines"`
}

