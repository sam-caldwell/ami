package types

// EventType models a builtin generic Event<T> type in the AMI type system.
type EventType struct{ Elem Type }

func (e EventType) String() string { return "Event<" + e.Elem.String() + ">" }

// ErrorType models a builtin generic Error<E> type in the AMI type system.
type ErrorType struct{ Elem Type }

func (e ErrorType) String() string { return "Error<" + e.Elem.String() + ">" }

// Drop and Ack represent control outcomes for worker results.
var (
    Drop = Basic{"drop"}
    Ack  = Basic{"ack"}
)

