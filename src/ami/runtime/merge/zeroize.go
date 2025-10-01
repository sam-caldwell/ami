package merge

// sensitiveZeroizer, when set, is called with an event payload about to be dropped
// from buffers (backpressure/expiry). It should aggressively overwrite sensitive
// data and return a count of zeroized fields/bytes for potential future metrics.
// This is a best-effort simulation in Go and is primarily used by tests.
var sensitiveZeroizer func(any) int

// SetSensitiveZeroizer installs a package-level zeroizer invoked on buffer drops.
// Passing nil disables zeroization. This is intended for tests and controlled
// environments; production zeroization happens at the codegen/runtime layer.
func SetSensitiveZeroizer(fn func(any) int) { sensitiveZeroizer = fn }

func zeroizePayload(p any) {
    if sensitiveZeroizer != nil { _ = sensitiveZeroizer(p) }
}

