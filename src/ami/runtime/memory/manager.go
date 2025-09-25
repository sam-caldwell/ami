package memory

import "sync"

// Domain identifies an allocation domain in the AMI memory model.
// Event: per-event heap; State: node-state heap; Ephemeral: call-local.
type Domain string

const (
	Event     Domain = "event"
	State     Domain = "state"
	Ephemeral Domain = "ephemeral"
)

// Manager provides a per-VM memory manager scaffold with simple accounting.
// It is not a real allocator; it tracks allocations by domain and supports
// deterministic release via RAII-like Handles.
type Manager struct {
	mu       sync.Mutex
	counters map[Domain]int
}

// NewManager creates a new memory Manager with zeroed counters.
func NewManager() *Manager {
	return &Manager{counters: map[Domain]int{Event: 0, State: 0, Ephemeral: 0}}
}

// Handle represents an allocated unit within a domain which must be released.
type Handle struct {
	m    *Manager
	dom  Domain
	n    int
	once sync.Once
}

// Alloc accounts for n units in the specified domain and returns a Handle
// whose Release method decrements the counter exactly once.
func (m *Manager) Alloc(dom Domain, n int) Handle {
	m.mu.Lock()
	m.counters[dom] += n
	m.mu.Unlock()
	return Handle{m: m, dom: dom, n: n}
}

// Release decrements the allocation counter exactly once.
func (h *Handle) Release() {
	if h.m == nil || h.n == 0 {
		return
	}
	h.once.Do(func() {
		h.m.mu.Lock()
		h.m.counters[h.dom] -= h.n
		h.m.mu.Unlock()
	})
}

// Stats returns a snapshot of allocation counters per domain.
func (m *Manager) Stats() map[Domain]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[Domain]int, len(m.counters))
	for k, v := range m.counters {
		out[k] = v
	}
	return out
}
