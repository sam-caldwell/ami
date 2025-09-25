package types

// EventType models a builtin generic Event<T> type in the AMI type system.
type EventType struct{ Elem Type }

func (e EventType) String() string { return "Event<" + e.Elem.String() + ">" }

// ErrorType models a builtin generic Error<E> type in the AMI type system.
type ErrorType struct{ Elem Type }

func (e ErrorType) String() string { return "Error<" + e.Elem.String() + ">" }

// Note: control outcomes like Ack/Drop are not part of the AMI type system.
