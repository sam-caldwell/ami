package memory

import "sync"

// Manager provides a per-VM memory manager scaffold with simple accounting.
// It is not a real allocator; it tracks allocations by domain and supports
// deterministic release via RAII-like Handles.
type Manager struct {
    mu       sync.Mutex
    counters map[Domain]int
}

