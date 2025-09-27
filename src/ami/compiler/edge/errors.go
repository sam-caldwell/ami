package edge

import "errors"

// ErrFull is returned by Push() when backpressure=block and the buffer is full.
var ErrFull = errors.New("edge full")

