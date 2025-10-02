package os

// internal helpers and errors
var (
    errInvalidProcess = &procError{"invalid process"}
    errAlreadyStarted = &procError{"process already started"}
)

type procError struct{ s string }
func (e *procError) Error() string { return e.s }

