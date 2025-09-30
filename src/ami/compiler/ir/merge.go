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
    LatePolicy  string      `json:"latePolicy,omitempty"` // drop|accept (docx governs)
}

type BufferPlan struct {
    Capacity int    `json:"capacity,omitempty"`
    Policy   string `json:"policy,omitempty"` // block|dropOldest|dropNewest
}

type SortKey struct { Field string `json:"field"`; Order string `json:"order,omitempty"` }

type Watermark struct { Field string `json:"field"`; LatenessMs int `json:"latenessMs,omitempty"` }
