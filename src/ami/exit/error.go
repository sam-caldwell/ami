package exit

import "fmt"

// Error wraps an error with an exit code.
type Error struct {
    Code Code
    Msg  string
}

func (e Error) Error() string { return e.Msg }

// New creates a new Error with the given code and formatted message.
func New(code Code, format string, a ...any) error {
    return Error{Code: code, Msg: fmt.Sprintf(format, a...)}
}

// UnwrapCode returns the exit code from an error, defaulting to Internal.
func UnwrapCode(err error) Code {
    if err == nil {
        return OK
    }
    if e, ok := err.(Error); ok {
        return e.Code
    }
    return Internal
}

