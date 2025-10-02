package exit

import "fmt"

// New creates a new Error with the given code and formatted message.
func New(code Code, format string, a ...any) error {
    return Error{Code: code, Msg: fmt.Sprintf(format, a...)}
}
