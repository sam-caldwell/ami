package logschema

// Level for log.v1 records. Matches logger levels.
type Level string

const (
    Trace Level = "trace"
    Debug Level = "debug"
    Info  Level = "info"
    Warn  Level = "warn"
    Error Level = "error"
    Fatal Level = "fatal"
)

