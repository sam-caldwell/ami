package schemas

// MergeConfigV1 captures normalized merge attributes for edge.MultiPath on Collect.
// All fields are optional and reflect tolerant parsing; when absent, defaults apply.
type MergeConfigV1 struct {
    SortField          string `json:"sortField,omitempty"`
    SortOrder          string `json:"sortOrder,omitempty"` // asc|desc
    Stable             bool   `json:"stable,omitempty"`
    Key                string `json:"key,omitempty"`
    Dedup              bool   `json:"dedup,omitempty"`
    DedupField         string `json:"dedupField,omitempty"`
    Window             int    `json:"window,omitempty"`
    WatermarkField     string `json:"watermarkField,omitempty"`
    WatermarkLateness  string `json:"watermarkLateness,omitempty"` // tolerant units string
    TimeoutMs          int    `json:"timeoutMs,omitempty"`
    BufferCapacity     int    `json:"bufferCapacity,omitempty"`
    BufferBackpressure string `json:"bufferBackpressure,omitempty"` // block|dropOldest|dropNewest
    PartitionBy        string `json:"partitionBy,omitempty"`
}

