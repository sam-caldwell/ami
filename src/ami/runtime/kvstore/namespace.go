package kvstore

import "sync"

// Registry tracks namespaced stores by key (e.g., pipeline/node).
type Registry struct {
    mu     sync.RWMutex
    stores map[string]*Store
}

// default registry
var defaultRegistry = &Registry{stores: map[string]*Store{}}

