package memory

// Alloc accounts for n units in the specified domain and returns a Handle
// whose Release method decrements the counter exactly once.
func (m *Manager) Alloc(dom Domain, n int) Handle {
    m.mu.Lock()
    m.counters[dom] += n
    m.mu.Unlock()
    return Handle{m: m, dom: dom, n: n}
}

