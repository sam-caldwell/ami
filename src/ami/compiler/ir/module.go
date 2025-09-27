package ir

// Module groups functions for a package.
type Module struct {
    Package   string
    Functions []Function
    Directives []Directive
    Concurrency int    `json:"concurrency,omitempty"`
    Backpressure string `json:"backpressurePolicy,omitempty"`
    TelemetryEnabled bool `json:"telemetryEnabled,omitempty"`
    Capabilities []string `json:"capabilities,omitempty"`
    TrustLevel   string   `json:"trustLevel,omitempty"`
}
