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
    stdinBuf  bytes.Buffer   // buffers pre-start stdin writes
    stdoutBuf bytes.Buffer
    stderrBuf bytes.Buffer

    mu       sync.Mutex
    started  bool
    exitCode *int
    preClosed bool // whether pre-start stdin writer was closed
}

// Start launches the process. When block is true, it waits for completion
// and records the exit code; otherwise it returns immediately and records
// exit code asynchronously when the process exits.
func (p *Process) Start(block bool) error {
    if p == nil || p.cmd == nil { return errInvalidProcess }
    p.mu.Lock()
    if p.started { p.mu.Unlock(); return errAlreadyStarted }
    p.started = true
    // Build stdin reader: buffered pre-start data followed by live pipe for post-start writes
    pr, pw := io.Pipe()
    // expose writer for post-start writes
    p.stdin = pw
    // initialize Stdin as multi-reader
    pre := bytes.NewReader(p.stdinBuf.Bytes())
    p.cmd.Stdin = io.NopCloser(io.MultiReader(pre, pr))
    // If blocking and pre-start writer was closed (no further input), close live pipe to signal EOF
    _ = p.preClosed
    p.mu.Unlock()
    if block {
        // In blocking mode, close the live pipe immediately to signal EOF
        // to the child's stdin copier, avoiding Run() blocking on stdin.
        _ = pw.Close()
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

func (p *Process) setExitCode(c int) { p.mu.Lock(); defer p.mu.Unlock(); p.exitCode = &c }

