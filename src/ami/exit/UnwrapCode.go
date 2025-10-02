package exit

import (
	"errors"
)

// UnwrapCode returns the exit code from an error, defaulting to Internal.
func UnwrapCode(err error) Code {
	if err == nil {
		return OK
	}
	var e Error
	if errors.As(err, &e) {
		return e.Code
	}
	return Internal
}
