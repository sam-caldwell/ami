package ir

type CollectSpec struct {
    Step  string     `json:"step"`
    Merge *MergePlan `json:"merge,omitempty"`
}

