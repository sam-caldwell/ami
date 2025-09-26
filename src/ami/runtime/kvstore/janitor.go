package kvstore

import "time"

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

