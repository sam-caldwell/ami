package kvstore

// Stats holds counters and gauges for a Store.
type Stats struct {
    Hits        uint64
    Misses      uint64
    Expirations uint64
    Evictions   uint64
    Entries     int
    BytesUsed   int64
}

