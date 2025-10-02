package ir

// Pipeline describes a pipeline IR projection sufficient to communicate merge configuration.
type Pipeline struct {
    Name    string        `json:"name"`
    Collect []CollectSpec `json:"collect,omitempty"`
}

