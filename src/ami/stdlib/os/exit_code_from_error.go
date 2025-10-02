package os

import "os/exec"

// Avoid importing os/syscall for a simple exit code extraction; exec already
// stores ProcessState.ExitCode where available. Fallback: try to read from
// returned error string when possible; otherwise nil.
func exitCodeFromError(err error) *int {
    if err == nil { return nil }
    if ee, ok := err.(*exec.ExitError); ok && ee.ProcessState != nil {
        c := ee.ProcessState.ExitCode()
        return &c
    }
    return nil
}

