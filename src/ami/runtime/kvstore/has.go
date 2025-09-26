package kvstore

import "time"

// Has reports whether a key exists and is not expired (lazily purging if expired).
func (s *Store) Has(key string) bool {
    now := time.Now()
    s.mu.Lock()
    defer s.mu.Unlock()
    if e, ok := s.items[key]; ok {
        if s.isExpired(e, now) {
            s.expireLocked(key, e)
            return false
        }
        return true
    }
    return false
}

