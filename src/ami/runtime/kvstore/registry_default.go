package kvstore

import "time"

// Default returns a process-wide registry suitable for scaffolding.
// Capacity is unlimited; TTL sweep runs at 100ms.
func Default() *Registry {
    defaultOnce.Do(func() {
        defaultReg = NewRegistry(Options{CapacityBytes: 0, SweepInterval: 100 * time.Millisecond})
    })
    return defaultReg
}

