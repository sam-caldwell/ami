package logger

// Sink is a minimal sink interface for Phase 1.
// Start initializes resources (e.g., open files). Write appends a single line.
// Close releases resources. Implementations should be safe for single-goroutine tests.
type Sink interface {
    Start() error
    Write(line []byte) error
    Close() error
}

