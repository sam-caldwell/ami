package edge

// Event is a minimal AMI event scaffold used by edge queues in tests/runtime
// harness. In the language, Event<T> is a builtin generic type; this Go type
// mirrors that shape minimally for push/pop validation (payload is untyped).
type Event struct {
	Payload any
}
