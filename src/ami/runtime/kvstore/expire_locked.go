package kvstore

// internal: expire e for key k, update metrics.
func (s *Store) expireLocked(k string, e *entry) {
    s.deleteLocked(k, e)
    s.expirations++
}

