package logging

// Level represents the log severity level.
// Values are lowercase to match JSON schema deterministically.
type Level string

const (
    LevelTrace Level = "trace"
    LevelDebug Level = "debug"
    LevelInfo  Level = "info"
    LevelWarn  Level = "warn"
    LevelError Level = "error"
    LevelFatal Level = "fatal"
)

