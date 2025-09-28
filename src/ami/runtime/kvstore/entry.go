package kvstore

import "time"

// entry stores a value and its expiration/read-count metadata.
type entry struct {
    val            any
    expireAt       time.Time
    remainingReads int
}

func (e *entry) isExpiredAt(t time.Time) bool {
    return !e.expireAt.IsZero() && !t.Before(e.expireAt)
}

func (e *entry) expired() bool { return e.isExpiredAt(time.Now()) }

func (e *entry) expiredLocked() bool { return e.expired() }

