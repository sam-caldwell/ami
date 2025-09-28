package kvstore

// Metrics captures store counters for observability.
type Metrics struct {
    Hits        int
    Misses      int
    Expirations int
    Evictions   int
    CurrentSize int
}
