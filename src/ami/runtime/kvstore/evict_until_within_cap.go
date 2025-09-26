package kvstore

// internal: evict until under capacity.
func (s *Store) evictUntilWithinCap() {
    if s.capBytes <= 0 {
        return
    }
    for s.used > s.capBytes {
        back := s.lru.Back()
        if back == nil {
            break
        }
        e, _ := back.Value.(*entry)
        if e == nil {
            break
        }
        delete(s.items, e.key)
        s.lru.Remove(back)
        delete(s.index, e.key)
        s.used -= e.size
        s.evictions++
    }
}

