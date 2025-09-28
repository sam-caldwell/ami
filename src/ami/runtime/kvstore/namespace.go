package kvstore

import "sync"

// Registry tracks namespaced stores by key (e.g., pipeline/node).
type Registry struct {
    mu     sync.RWMutex
    stores map[string]*Store
}

// default registry
var defaultRegistry = &Registry{stores: map[string]*Store{}}

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

// ResetRegistry resets the default registry (for tests).
func ResetRegistry() {
    defaultRegistry = &Registry{stores: map[string]*Store{}}
}

