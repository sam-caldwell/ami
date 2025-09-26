package kvstore

import "time"

// Put inserts or replaces the value for key with optional TTL and read-count policy.
func (s *Store) Put(key string, val any, opts ...PutOption) {
    cfg := putConfig{}
    for _, o := range opts {
        o(&cfg)
    }
    now := time.Now()
    // Compute approx size
    sz := int64(len(key)) + approxSize(val)

    s.mu.Lock()
    defer s.mu.Unlock()

    // Replace existing if present
    if e, ok := s.items[key]; ok {
        s.used -= e.size
        s.lru.Remove(e.elem)
        delete(s.index, key)
    }

    e := &entry{
        key:            key,
        value:          val,
        size:           sz,
        readsRemaining: cfg.maxReads,
        lastAccess:     now,
    }
    if cfg.ttl > 0 {
        e.ttl = cfg.ttl
        e.sliding = cfg.sliding
        e.expiresAt = now.Add(cfg.ttl)
    }

    // Insert and update LRU
    elem := s.lru.PushFront(e)
    e.elem = elem
    s.index[key] = elem
    s.items[key] = e
    s.used += e.size

    // Enforce capacity via LRU eviction
    s.evictUntilWithinCap()
}

