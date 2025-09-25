package kvstore

import (
	"fmt"
	"sync"
	"time"
)

// Registry provides namespaced Store instances per (pipeline, node) tuple.
type Registry struct {
	mu    sync.Mutex
	opts  Options
	table map[string]*Store
}

// NewRegistry creates a new Registry with Store options used for new instances.
func NewRegistry(opts Options) *Registry {
	return &Registry{opts: opts, table: make(map[string]*Store)}
}

func key(pipeline, node string) string { return fmt.Sprintf("%s\x1f%s", pipeline, node) }

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

// Close stops background janitors for all registered stores.
func (r *Registry) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.table {
		s.Close()
	}
}

// Default global registry with sensible defaults for tests/runtime scaffolding.
var (
	defaultOnce sync.Once
	defaultReg  *Registry
	defaultMu   sync.Mutex
)

// Default returns a process-wide registry suitable for scaffolding.
// Capacity is unlimited; TTL sweep runs at 100ms.
func Default() *Registry {
	defaultOnce.Do(func() {
		defaultReg = NewRegistry(Options{CapacityBytes: 0, SweepInterval: 100 * time.Millisecond})
	})
	return defaultReg
}

// Reset clears the registry, closing all stores and removing them.
func (r *Registry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.table {
		s.Close()
	}
	r.table = make(map[string]*Store)
}

// ResetDefault replaces the process-wide default registry with a fresh instance.
// An optional Options value may be provided; otherwise defaults are used.
func ResetDefault(opts ...Options) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultReg != nil {
		defaultReg.Close()
	}
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	} else {
		o = Options{CapacityBytes: 0, SweepInterval: 100 * time.Millisecond}
	}
	defaultReg = NewRegistry(o)
}

// StoreInfo captures a snapshot of a registered store for observability.
type StoreInfo struct {
	Pipeline string
	Node     string
	Stats    Stats
	Dump     string
}

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
