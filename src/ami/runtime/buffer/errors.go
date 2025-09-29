package buffer

import "errors"

// ErrFull indicates the buffer is at capacity for block backpressure.
var ErrFull = errors.New("buffer full")

