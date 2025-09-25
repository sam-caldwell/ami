package bufio

import "errors"

// ErrInvalidBufferSize indicates a non-positive buffer size.
var ErrInvalidBufferSize = errors.New("invalid buffer size; must be > 0")

