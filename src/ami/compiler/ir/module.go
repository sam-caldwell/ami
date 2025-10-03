package ir

// Module groups functions for a package.
type Module struct {
    Package   string
    Functions []Function
    Directives []Directive
    Pipelines []Pipeline `json:"pipelines,omitempty"`
    // ErrorPipes lists per-pipeline error routes captured from the frontend.
    // This enables backends to embed metadata for error handling paths.
    ErrorPipes []ErrorPipeline `json:"errorPipelines,omitempty"`
    Concurrency int    `json:"concurrency,omitempty"`
    Backpressure string `json:"backpressurePolicy,omitempty"`
    Schedule    string `json:"schedule,omitempty"`
    TelemetryEnabled bool `json:"telemetryEnabled,omitempty"`
    Capabilities []string `json:"capabilities,omitempty"`
    TrustLevel   string   `json:"trustLevel,omitempty"`
    ExecContext *ExecContext `json:"execContext,omitempty"`
    EventMeta  *EventMeta   `json:"eventmeta,omitempty"`
}
