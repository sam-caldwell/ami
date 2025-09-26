package kvstore

import (
    "container/list"
    "time"
)

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

