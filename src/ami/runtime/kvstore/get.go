package kvstore

import "time"

// Get retrieves the value if present and not expired. It updates LRU and metrics.
func (s *Store) Get(key string) (any, bool) {
    now := time.Now()
    s.mu.Lock()
    defer s.mu.Unlock()

    e, ok := s.items[key]
    if !ok {
        s.misses++
        return nil, false
    }
    if s.isExpired(e, now) {
        // account expiration on access
        s.expireLocked(key, e)
        s.misses++
        return nil, false
    }
    // Sliding TTL refresh
    if e.sliding && e.ttl > 0 {
        e.expiresAt = now.Add(e.ttl)
    }

    // Delete-on-read after N reads
    shouldDelete := false
    if e.readsRemaining > 0 {
        if e.readsRemaining == 1 {
            shouldDelete = true
        }
        e.readsRemaining--
    }

    // Update LRU
    e.lastAccess = now
    s.lru.MoveToFront(e.elem)
    s.hits++

    val := e.value
    if shouldDelete {
        s.deleteLocked(key, e)
    }
    return val, true
}

