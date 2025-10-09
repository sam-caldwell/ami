package bufio

import "errors"

var (
    ErrInvalidArg     = errors.New("bufio: invalid argument")
    ErrInvalidHandle  = errors.New("bufio: invalid handle")
    ErrAlreadyReleased = errors.New("bufio: already released")
)

