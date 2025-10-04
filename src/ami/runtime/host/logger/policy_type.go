package logger

import "errors"

// BackpressurePolicy selects behavior when the queue is full.
type BackpressurePolicy string

const (
    Block      BackpressurePolicy = "block"
    DropNewest BackpressurePolicy = "dropNewest"
    DropOldest BackpressurePolicy = "dropOldest"
)

var ErrDropped = errors.New("dropped due to backpressure policy")

