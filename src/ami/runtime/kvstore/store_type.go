package kvstore

import (
    "container/list"
    "sync"
    "time"
)

// Store is a concurrency-safe ephemeral key/value store with TTL and LRU eviction.
type Store struct {
    mu       sync.Mutex
    items    map[string]*entry
    lru      *list.List // front: most recent; back: least recent
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

