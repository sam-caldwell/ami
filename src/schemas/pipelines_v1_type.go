package schemas

// PipelinesV1 is a debug schema that captures lowered pipeline structure
// for a compilation unit, including referenced workers and generic payloads.
// Emitted under build/debug/ir/<package>/<unit>.pipelines.json when verbose build is enabled.
type PipelinesV1 struct {
    Schema    string       `json:"schema"`
    Timestamp string       `json:"timestamp"`
    Package   string       `json:"package"`
    File      string       `json:"file"`
    Pipelines []PipelineV1 `json:"pipelines"`
}

