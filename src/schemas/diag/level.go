package diag

// Level is the diagnostic severity for diag.v1.
// Only info, warn, and error are valid levels.
type Level string

const (
    Info  Level = "info"
    Warn  Level = "warn"
    Error Level = "error"
)

