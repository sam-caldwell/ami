package kvstore

// Snapshot returns a slice of StoreInfo for all registered stores.
func (r *Registry) Snapshot() []StoreInfo {
    r.mu.Lock()
    defer r.mu.Unlock()
    out := make([]StoreInfo, 0, len(r.table))
    for k, s := range r.table {
        // split key back into components: pipeline\x1fnode
        var pipeline, node string
        for i := 0; i < len(k); i++ {
            if k[i] == '\x1f' {
                pipeline, node = k[:i], k[i+1:]
                break
            }
        }
        if node == "" {
            pipeline = k
        }
        out = append(out, StoreInfo{Pipeline: pipeline, Node: node, Stats: s.Metrics(), Dump: s.DebugDump()})
    }
    return out
}

