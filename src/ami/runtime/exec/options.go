package exec

import "time"

// ExecOptions control runner behavior for sources/sinks in simulation.
type ExecOptions struct {
    SourceType    string        // auto|file|timer
    TimerInterval time.Duration // used when SourceType=timer or auto+Timer node
    TimerCount    int           // number of timer events (0=unlimited)
}

