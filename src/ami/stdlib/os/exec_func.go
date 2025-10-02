package os

import "os/exec"

// Exec initializes a command but does not start it. The returned Process
// can be started with Start(block).
func Exec(program string, args ...string) (*Process, error) {
    c := exec.Command(program, args...)
    p := &Process{cmd: c}
    // Wire stdout/stderr buffers for capture
    c.Stdout = &p.stdoutBuf
    c.Stderr = &p.stderrBuf
    // Pre-wire stdin with a buffer-backed writer. On Start, we will feed
    // the buffered bytes first, then a live pipe for post-start writes.
    p.stdin = &prestartWriter{p: p}
    // Apply platform process attributes (e.g., setpgid on Unix)
    applySysProcAttr(c)
    return p, nil
}

