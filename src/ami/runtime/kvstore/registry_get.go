package kvstore

// Get returns the Store instance for (pipeline,node), creating it if needed.
func (r *Registry) Get(pipeline, node string) *Store {
    k := key(pipeline, node)
    r.mu.Lock()
    defer r.mu.Unlock()
    if s, ok := r.table[k]; ok {
        return s
    }
    s := New(r.opts)
    r.table[k] = s
    return s
}

