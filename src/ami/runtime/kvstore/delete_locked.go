package kvstore

// internal: delete key and remove from LRU, adjust size.
func (s *Store) deleteLocked(k string, e *entry) {
    delete(s.items, k)
    if e.elem != nil {
        s.lru.Remove(e.elem)
    }
    delete(s.index, k)
    s.used -= e.size
}

