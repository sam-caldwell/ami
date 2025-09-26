package tester

// Result captures outcome for a single case.
type Result struct {
    Name       string
    Status     string // pass|fail|skip
    DurationMs int64
    Error      string // when fail/skip, a description
}

