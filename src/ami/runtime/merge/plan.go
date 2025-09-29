package merge

// Plan is a normalized configuration for a merge operator.
type Plan struct {
    Buffer struct{
        Capacity int
        Policy string // block|dropOldest|dropNewest|shuntOldest|shuntNewest
    }
    Stable bool
    Sort []SortKey
    Key string
    PartitionBy string
    Dedup struct{ Field string }
    Watermark *Watermark // nil means disabled
    TimeoutMs int // 0 disabled
    Window int // 0 disabled
}

type SortKey struct{ Field string; Order string /*asc|desc*/ }

type Watermark struct{ Field string; LatenessMs int }

