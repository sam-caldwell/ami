package logger

import "sync"

// Logger manages output formatting and routing for human and JSON logs.
type Logger struct {
    json    bool
    verbose bool
    color   bool
    mu      sync.Mutex
}

var std = &Logger{}

