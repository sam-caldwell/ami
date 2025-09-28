package kvstore

import (
    "sync"
    "time"
)

// Store is an in-memory ephemeral key/value store with TTL and delete-on-read.
// It is concurrency-safe and intended for per-node usage.
type Store struct {
    mu      sync.RWMutex
    items   map[string]*entry
    metrics Metrics
}

// New creates a new empty Store.
func New() *Store { return &Store{items: make(map[string]*entry)} }

// Put stores a value with optional TTL and maxReads (delete-on-read after N gets).
func (s *Store) Put(key string, val any, opts ...PutOption) {
    o := applyOptions(opts)
    s.mu.Lock()
    defer s.mu.Unlock()
    e := &entry{val: val}
    if o.ttl > 0 { e.expireAt = time.Now().Add(o.ttl) }
    if o.maxReads > 0 { e.remainingReads = o.maxReads }
    s.items[key] = e
}

// Get returns a value if present and not expired. It decrements remainingReads
// and deletes the key when remainingReads reaches zero.
func (s *Store) Get(key string) (any, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    e, ok := s.items[key]
    if !ok { s.metrics.Misses++; return nil, false }
    if e.expiredLocked() { delete(s.items, key); s.metrics.Expirations++; s.metrics.Misses++; return nil, false }
    v := e.val
    if e.remainingReads > 0 {
        e.remainingReads--
        if e.remainingReads == 0 { delete(s.items, key) }
    }
    s.metrics.Hits++
    return v, true
}

// Del deletes a key if it exists and returns true when removed.
func (s *Store) Del(key string) bool {
    s.mu.Lock()
    defer s.mu.Unlock()
    if _, ok := s.items[key]; ok { delete(s.items, key); return true }
    return false
}

// Has returns true if a non-expired key exists.
func (s *Store) Has(key string) bool {
    s.mu.RLock()
    e, ok := s.items[key]
    s.mu.RUnlock()
    if !ok { return false }
    if e.expired() { s.mu.Lock(); delete(s.items, key); s.mu.Unlock(); return false }
    return true
}

// Keys returns a snapshot of existing non-expired keys.
func (s *Store) Keys() []string {
    now := time.Now()
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := make([]string, 0, len(s.items))
    for k, e := range s.items {
        if !e.isExpiredAt(now) { out = append(out, k) }
    }
    return out
}

// Metrics returns a copy of current metrics.
func (s *Store) Metrics() Metrics {
    s.mu.RLock(); defer s.mu.RUnlock()
    m := s.metrics
    m.CurrentSize = len(s.items)
    return m
}

