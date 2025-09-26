package kvstore

// Close stops background janitors for all registered stores.
func (r *Registry) Close() {
    r.mu.Lock()
    defer r.mu.Unlock()
    for _, s := range r.table {
        s.Close()
    }
}

