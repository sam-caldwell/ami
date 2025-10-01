package os

import (
    "bytes"
    "io"
    "os/exec"
    "sync"
)

// Process represents a spawned process handle.
// It wraps exec.Cmd and exposes lifecycle and I/O accessors.
type Process struct {
    cmd       *exec.Cmd
    stdin     ioWriteCloser // lightweight alias to avoid importing io in public type
    stdoutBuf bytes.Buffer
    stderrBuf bytes.Buffer

    mu       sync.Mutex
    started  bool
    exitCode *int
}

// Exec initializes a command but does not start it. The returned Process
// can be started with Start(block).
func Exec(program string, args ...string) (*Process, error) {
    c := exec.Command(program, args...)
    p := &Process{cmd: c}
    // Wire stdout/stderr buffers for capture
    c.Stdout = &p.stdoutBuf
    c.Stderr = &p.stderrBuf
    // Pre-wire stdin with a pipe so Stdin() can be used before or after Start()
    pr, pw := io.Pipe()
    c.Stdin = pr
    p.stdin = pw
    // Apply platform process attributes (e.g., setpgid on Unix)
    applySysProcAttr(c)
    return p, nil
}

// Start launches the process. When block is true, it waits for completion
// and records the exit code; otherwise it returns immediately and records
// exit code asynchronously when the process exits.
func (p *Process) Start(block bool) error {
    if p == nil || p.cmd == nil { return errInvalidProcess }
    p.mu.Lock()
    if p.started { p.mu.Unlock(); return errAlreadyStarted }
    p.started = true
    p.mu.Unlock()
    if block {
        // Close stdin writer to avoid childStdin goroutine blocking during Run when no input is required.
        if p.stdin != nil { _ = p.stdin.Close() }
        if err := p.cmd.Run(); err != nil {
            if ee := exitCodeFromError(err); ee != nil { p.setExitCode(*ee) }
            return err
        }
        if p.cmd.ProcessState != nil { ec := p.cmd.ProcessState.ExitCode(); p.setExitCode(ec) }
        return nil
    }
    if err := p.cmd.Start(); err != nil { return err }
    // Wait in background to capture exit code
    go func(){ _ = p.cmd.Wait(); if p.cmd.ProcessState != nil { ec := p.cmd.ProcessState.ExitCode(); p.setExitCode(ec) } }()
    return nil
}

// Kill terminates the process immediately.
func (p *Process) Kill() error {
    if p == nil || p.cmd == nil || p.cmd.Process == nil { return errInvalidProcess }
    // Attempt to kill the entire process group when supported (e.g., Unix setpgid)
    if err := killProcessGroup(p.cmd.Process); err == nil { return nil }
    return p.cmd.Process.Kill()
}

// Pid returns the OS process id after Start. Before Start it returns 0.
func (p *Process) Pid() int {
    if p == nil || p.cmd == nil || p.cmd.Process == nil { return 0 }
    return p.cmd.Process.Pid
}

// Status returns a snapshot of process status.
func (p *Process) Status() ProcessStatus {
    p.mu.Lock(); defer p.mu.Unlock()
    ps := ProcessStatus{PID: p.Pid()}
    if !p.started {
        ps.Running = false
        return ps
    }
    if p.exitCode == nil {
        ps.Running = true
        return ps
    }
    ps.Running = false
    v := *p.exitCode
    ps.ExitCode = &v
    return ps
}

// Stdout returns all bytes written to stdout so far.
func (p *Process) Stdout() []byte { if p == nil { return nil }; return p.stdoutBuf.Bytes() }

// Stderr returns all bytes written to stderr so far.
func (p *Process) Stderr() []byte { if p == nil { return nil }; return p.stderrBuf.Bytes() }

// Stdin returns a writer to the process stdin. It creates the pipe upon first call.
func (p *Process) Stdin() (ioWriteCloser, error) {
    if p == nil || p.cmd == nil { return nil, errInvalidProcess }
    p.mu.Lock(); defer p.mu.Unlock()
    return p.stdin, nil
}

// ProcessStatus captures process runtime state.
type ProcessStatus struct {
    PID      int
    Running  bool
    ExitCode *int
}

func (p *Process) setExitCode(c int) { p.mu.Lock(); defer p.mu.Unlock(); p.exitCode = &c }

// internal helpers and errors
var (
    errInvalidProcess = &procError{"invalid process"}
    errAlreadyStarted = &procError{"process already started"}
)

type procError struct{ s string }
func (e *procError) Error() string { return e.s }

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

// Minimal interface to avoid pulling io in public signature; concrete value is io.WriteCloser.
type ioWriteCloser interface{ Write([]byte) (int, error); Close() error }
