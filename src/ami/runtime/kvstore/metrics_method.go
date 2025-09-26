package kvstore

// Metrics returns a snapshot of current store metrics and gauges.
func (s *Store) Metrics() Stats {
    s.mu.Lock()
    defer s.mu.Unlock()
    return Stats{
        Hits:        s.hits,
        Misses:      s.misses,
        Expirations: s.expirations,
        Evictions:   s.evictions,
        Entries:     len(s.items),
        BytesUsed:   s.used,
    }
}

