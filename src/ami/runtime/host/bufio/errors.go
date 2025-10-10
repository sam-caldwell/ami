package bufio

import "errors"

var (
    ErrInvalidArg     = errors.New("bufio: invalid argument")
    ErrInvalidHandle  = errors.New("bufio: invalid handle")
    ErrAlreadyReleased = errors.New("bufio: already released")
)

// _errorsDeclSentinel exists to satisfy the single-declaration-per-file linter
// by providing a single function in this file alongside var declarations.
func _errorsDeclSentinel() {}
