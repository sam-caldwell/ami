package memory

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

