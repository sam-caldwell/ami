package kvstore

import "sync"

// Registry provides namespaced Store instances per (pipeline, node) tuple.
type Registry struct {
    mu    sync.Mutex
    opts  Options
    table map[string]*Store
}

