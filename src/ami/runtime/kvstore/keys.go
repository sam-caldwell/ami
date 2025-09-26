package kvstore

import "time"

// Keys returns current keys after purging any expirations.
func (s *Store) Keys() []string {
    now := time.Now()
    s.mu.Lock()
    defer s.mu.Unlock()
    // Purge expired lazily
    for k, e := range s.items {
        if s.isExpired(e, now) {
            s.expireLocked(k, e)
        }
    }
    out := make([]string, 0, len(s.items))
    for k := range s.items {
        out = append(out, k)
    }
    return out
}

