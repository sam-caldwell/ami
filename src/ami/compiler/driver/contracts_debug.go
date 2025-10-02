package driver

type contractsDoc struct {
    Schema       string               `json:"schema"`
    Package      string               `json:"package"`
    Unit         string               `json:"unit"`
    Delivery     string               `json:"delivery"`
    Capabilities []string             `json:"capabilities,omitempty"`
    TrustLevel   string               `json:"trustLevel,omitempty"`
    Concurrency  *contractConcurrency  `json:"concurrency,omitempty"`
    Pipelines    []contractPipeline   `json:"pipelines"`
    CapabilityNotes []string          `json:"capabilityNotes,omitempty"`
    SchemaStability string            `json:"schemaStability,omitempty"`
    SchemaNotes   []string            `json:"schemaNotes,omitempty"`
}
// pipeline/step/concurrency types and writer moved to separate files
