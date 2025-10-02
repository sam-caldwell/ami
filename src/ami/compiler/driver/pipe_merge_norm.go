package driver

type pipeMergeNorm struct {
    Buffer *struct{
        Capacity int    `json:"capacity"`
        Policy   string `json:"policy,omitempty"`
    } `json:"buffer,omitempty"`
    Stable bool `json:"stable,omitempty"`
    Sort   []struct{
        Field string `json:"field"`
        Order string `json:"order"`
    } `json:"sort,omitempty"`
    Key         string `json:"key,omitempty"`
    PartitionBy string `json:"partitionBy,omitempty"`
    TimeoutMs   int    `json:"timeoutMs,omitempty"`
    Window      int    `json:"window,omitempty"`
    Watermark   *struct{
        Field    string `json:"field"`
        Lateness string `json:"lateness"`
    } `json:"watermark,omitempty"`
    Dedup       string `json:"dedup,omitempty"`
}

