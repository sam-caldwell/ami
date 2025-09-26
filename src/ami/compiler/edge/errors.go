package edge

import "errors"

// ErrFull indicates a bounded edge is full and would block producers.
var ErrFull = errors.New("edge buffer full (would block)")

