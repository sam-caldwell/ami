package kvstore

import "time"

// internal: determine expiry vs now.
func (s *Store) isExpired(e *entry, now time.Time) bool {
    if e.expiresAt.IsZero() {
        return false
    }
    return now.After(e.expiresAt)
}

