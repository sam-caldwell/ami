package diag

// Level represents the severity of a diagnostic message.
type Level string

const (
    Info  Level = "info"
    Warn  Level = "warn"
    Error Level = "error"
)

