package logger

import "time"

// Config configures a Pipeline over a Sink.
type Config struct {
    Capacity      int
    BatchMax      int
    FlushInterval time.Duration
    Policy        BackpressurePolicy
    Sink          Sink
    // Optional redaction for JSON log.v1 lines
    JSONRedactKeys     []string
    JSONRedactPrefixes []string
    JSONAllowKeys      []string
    JSONDenyKeys       []string
}

