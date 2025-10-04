package os

// ProcessStatus captures process runtime state.
type ProcessStatus struct {
    PID      int
    Running  bool
    ExitCode *int
}

