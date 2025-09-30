package ir

// Pipeline describes a pipeline IR projection sufficient to communicate
// Collect/merge configuration to the runtime.
type Pipeline struct {
    Name    string        `json:"name"`
    Collect []CollectSpec `json:"collect,omitempty"`
}

type CollectSpec struct {
    Step  string     `json:"step"`
    Merge *MergePlan `json:"merge,omitempty"`
}

