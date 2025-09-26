package kvstore

import (
    "container/list"
    "time"
)

// entry stores a value and its lifecycle controls.
type entry struct {
    key            string
    value          any
    size           int64 // approx bytes used (key+value)
    expiresAt      time.Time
    ttl            time.Duration
    sliding        bool
    readsRemaining int // <=0 means unlimited
    lastAccess     time.Time
    // lru element pointer maintained by Store
    elem *list.Element
}

