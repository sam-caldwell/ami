package kvstore

// Reset clears the registry, closing all stores and removing them.
func (r *Registry) Reset() {
    r.mu.Lock()
    defer r.mu.Unlock()
    for _, s := range r.table {
        s.Close()
    }
    r.table = make(map[string]*Store)
}

