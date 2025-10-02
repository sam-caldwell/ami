package merge

// sensitiveZeroizer, when set, is called with an event payload about to be dropped from buffers.
var sensitiveZeroizer func(any) int

func zeroizePayload(p any) { if sensitiveZeroizer != nil { _ = sensitiveZeroizer(p) } }

