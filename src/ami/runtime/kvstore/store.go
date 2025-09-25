package kvstore

import (
    "container/list"
    "encoding/json"
    "fmt"
    "sync"
    "time"
)

// Stats holds counters and gauges for a Store.
type Stats struct {
    Hits        uint64
    Misses      uint64
    Expirations uint64
    Evictions   uint64
    Entries     int
    BytesUsed   int64
}

// entry stores a value and its lifecycle controls.
type entry struct {
    key           string
    value         any
    size          int64 // approx bytes used (key+value)
    expiresAt     time.Time
    ttl           time.Duration
    sliding       bool
    readsRemaining int // <=0 means unlimited
    lastAccess    time.Time
    // lru element pointer maintained by Store
    elem *list.Element
}

// Store is a concurrency-safe ephemeral key/value store with TTL and LRU eviction.
type Store struct {
    mu       sync.Mutex
    items    map[string]*entry
    lru      *list.List          // front: most recent; back: least recent
    index    map[string]*list.Element
    capBytes int64
    used     int64

    // metrics
    hits        uint64
    misses      uint64
    expirations uint64
    evictions   uint64

    // janitor
    sweepEvery time.Duration
    stopCh     chan struct{}
    wg         sync.WaitGroup
}

// New creates a Store with the given options.
func New(opts Options) *Store {
    s := &Store{
        items:      make(map[string]*entry),
        lru:        list.New(),
        index:      make(map[string]*list.Element),
        capBytes:   opts.CapacityBytes,
        sweepEvery: opts.SweepInterval,
        stopCh:     make(chan struct{}),
    }
    if s.sweepEvery > 0 {
        s.wg.Add(1)
        go s.janitor()
    }
    return s
}

// Close stops background janitor (if running). The store remains usable.
func (s *Store) Close() {
    s.mu.Lock()
    if s.stopCh != nil {
        close(s.stopCh)
    }
    s.mu.Unlock()
    s.wg.Wait()
}

// Put inserts or replaces the value for key with optional TTL and read-count policy.
func (s *Store) Put(key string, val any, opts ...PutOption) {
    cfg := putConfig{}
    for _, o := range opts { o(&cfg) }
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
    if e.sliding && e.ttl > 0 { e.expiresAt = now.Add(e.ttl) }

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

// Del removes a key if present and returns true when deleted.
func (s *Store) Del(key string) bool {
    s.mu.Lock(); defer s.mu.Unlock()
    if e, ok := s.items[key]; ok {
        s.deleteLocked(key, e)
        return true
    }
    return false
}

// Has reports whether a key exists and is not expired (lazily purging if expired).
func (s *Store) Has(key string) bool {
    now := time.Now()
    s.mu.Lock(); defer s.mu.Unlock()
    if e, ok := s.items[key]; ok {
        if s.isExpired(e, now) {
            s.expireLocked(key, e)
            return false
        }
        return true
    }
    return false
}

// Keys returns current keys after purging any expirations.
func (s *Store) Keys() []string {
    now := time.Now()
    s.mu.Lock(); defer s.mu.Unlock()
    // Purge expired lazily
    for k, e := range s.items {
        if s.isExpired(e, now) {
            s.expireLocked(k, e)
        }
    }
    out := make([]string, 0, len(s.items))
    for k := range s.items { out = append(out, k) }
    return out
}

// Metrics returns a snapshot of current store metrics and gauges.
func (s *Store) Metrics() Stats {
    s.mu.Lock(); defer s.mu.Unlock()
    return Stats{
        Hits:        s.hits,
        Misses:      s.misses,
        Expirations: s.expirations,
        Evictions:   s.evictions,
        Entries:     len(s.items),
        BytesUsed:   s.used,
    }
}

// DebugDump returns a human-readable summary of the store's contents.
func (s *Store) DebugDump() string {
    s.mu.Lock(); defer s.mu.Unlock()
    type dumpEntry struct{
        Key string `json:"key"`
        ExpiresAt string `json:"expiresAt,omitempty"`
        Sliding bool `json:"sliding,omitempty"`
        ReadsRemaining int `json:"readsRemaining,omitempty"`
        Size int64 `json:"size"`
        LastAccess string `json:"lastAccess"`
    }
    tmp := struct{
        Entries int `json:"entries"`
        BytesUsed int64 `json:"bytesUsed"`
        Items []dumpEntry `json:"items"`
    }{Entries: len(s.items), BytesUsed: s.used}
    tmp.Items = make([]dumpEntry, 0, len(s.items))
    for _, e := range s.items {
        de := dumpEntry{Key: e.key, Size: e.size, ReadsRemaining: e.readsRemaining, LastAccess: e.lastAccess.Format(time.RFC3339Nano)}
        if !e.expiresAt.IsZero() { de.ExpiresAt = e.expiresAt.Format(time.RFC3339Nano) }
        if e.sliding { de.Sliding = true }
        tmp.Items = append(tmp.Items, de)
    }
    b, _ := json.Marshal(tmp)
    return string(b)
}

// internal: janitor sweeps TTL expirations periodically.
func (s *Store) janitor() {
    defer s.wg.Done()
    ticker := time.NewTicker(s.sweepEvery)
    defer ticker.Stop()
    for {
        select {
        case <-s.stopCh:
            return
        case now := <-ticker.C:
            s.mu.Lock()
            for k, e := range s.items {
                if s.isExpired(e, now) {
                    s.expireLocked(k, e)
                }
            }
            s.mu.Unlock()
        }
    }
}

// internal: determine expiry vs now.
func (s *Store) isExpired(e *entry, now time.Time) bool {
    if e.expiresAt.IsZero() { return false }
    return now.After(e.expiresAt)
}

// internal: expire e for key k, update metrics.
func (s *Store) expireLocked(k string, e *entry) {
    s.deleteLocked(k, e)
    s.expirations++
}

// internal: delete key and remove from LRU, adjust size.
func (s *Store) deleteLocked(k string, e *entry) {
    delete(s.items, k)
    if e.elem != nil { s.lru.Remove(e.elem) }
    delete(s.index, k)
    s.used -= e.size
}

// internal: evict until under capacity.
func (s *Store) evictUntilWithinCap() {
    if s.capBytes <= 0 { return }
    for s.used > s.capBytes {
        back := s.lru.Back()
        if back == nil { break }
        e, _ := back.Value.(*entry)
        if e == nil { break }
        delete(s.items, e.key)
        s.lru.Remove(back)
        delete(s.index, e.key)
        s.used -= e.size
        s.evictions++
    }
}

// approxSize computes an approximate size of v in bytes by JSON encoding,
// falling back to fmt.Sprintf when necessary.
func approxSize(v any) int64 {
    if v == nil { return 0 }
    switch t := v.(type) {
    case string:
        return int64(len(t))
    case []byte:
        return int64(len(t))
    default:
        if b, err := json.Marshal(t); err == nil { return int64(len(b)) }
        return int64(len(fmt.Sprintf("%v", v)))
    }
}

