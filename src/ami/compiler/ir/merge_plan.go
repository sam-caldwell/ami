package ir

// MergePlan encodes Collect/merge behavior in IR for downstream runtimes.
type MergePlan struct {
    Stable      bool        `json:"stable,omitempty"`
    Sort        []SortKey   `json:"sort,omitempty"`
    Key         string      `json:"key,omitempty"`
    PartitionBy string      `json:"partitionBy,omitempty"`
    Buffer      BufferPlan  `json:"buffer,omitempty"`
    Window      int         `json:"window,omitempty"`
    TimeoutMs   int         `json:"timeoutMs,omitempty"`
    DedupField  string      `json:"dedupField,omitempty"`
    Watermark   *Watermark  `json:"watermark,omitempty"`
    LatePolicy  string      `json:"latePolicy,omitempty"`
}

