package kvstore

// Del removes a key if present and returns true when deleted.
func (s *Store) Del(key string) bool {
    s.mu.Lock()
    defer s.mu.Unlock()
    if e, ok := s.items[key]; ok {
        s.deleteLocked(key, e)
        return true
    }
    return false
}

