package kvstore

// Namespace returns (and creates if missing) a store for the given namespace.
func Namespace(ns string) *Store {
    defaultRegistry.mu.RLock()
    s, ok := defaultRegistry.stores[ns]
    defaultRegistry.mu.RUnlock()
    if ok && s != nil { return s }
    defaultRegistry.mu.Lock()
    defer defaultRegistry.mu.Unlock()
    if s = defaultRegistry.stores[ns]; s != nil { return s }
    s = New()
    defaultRegistry.stores[ns] = s
    return s
}

