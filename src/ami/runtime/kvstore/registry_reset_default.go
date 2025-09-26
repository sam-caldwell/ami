package kvstore

import "time"

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

