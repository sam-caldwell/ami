package kvstore

import "sync"

// Default global registry with sensible defaults for tests/runtime scaffolding.
var (
    defaultOnce sync.Once
    defaultReg  *Registry
    defaultMu   sync.Mutex
)

