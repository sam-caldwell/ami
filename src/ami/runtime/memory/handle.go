package memory

import "sync"

// Handle represents an allocated unit within a domain which must be released.
type Handle struct {
    m    *Manager
    dom  Domain
    n    int
    once sync.Once
}

