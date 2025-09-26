package kvstore

// Close stops background janitor (if running). The store remains usable.
func (s *Store) Close() {
    s.mu.Lock()
    if s.stopCh != nil {
        close(s.stopCh)
    }
    s.mu.Unlock()
    s.wg.Wait()
}

