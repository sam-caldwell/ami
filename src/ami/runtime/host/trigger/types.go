package trigger

import (
    "errors"
    amitime "github.com/sam-caldwell/ami/src/ami/runtime/host/time"
)

// ErrNotImplemented indicates a feature is not yet available in the stdlib.
var ErrNotImplemented = errors.New("trigger: not implemented")

// Event is a simple generic wrapper carrying a value and a timestamp.
// This aligns with the spec's conceptual Event<T>.
type Event[T any] struct { Value T; Timestamp amitime.Time }
 
