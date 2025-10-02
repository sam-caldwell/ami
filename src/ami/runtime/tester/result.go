package tester

import "time"

// Result captures execution outcome for a single case.
type Result struct {
    Output   map[string]any
    ErrCode  string
    ErrMsg   string
    Duration time.Duration
}

