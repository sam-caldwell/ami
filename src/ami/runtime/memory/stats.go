package memory

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

