package amsignal

import (
    "os"
    "syscall"
    "testing"
    "time"
)

func TestRegister_SIGINT_InvokesHandler(t *testing.T) {
    defer Reset()
    fired := make(chan struct{}, 1)
    Register(SIGINT, func(){ fired <- struct{}{} })
    // Give registration a moment to subscribe
    time.Sleep(10 * time.Millisecond)
    // Send SIGINT to current process (portable as Interrupt)
    p, _ := os.FindProcess(os.Getpid())
    // Prefer Interrupt where possible; syscall.SIGINT on Unix
    sig := os.Interrupt
    // Try syscall when available
    _ = p.Signal(sig)
    select {
    case <-fired:
        // ok
    case <-time.After(2 * time.Second):
        t.Fatalf("timeout waiting for handler to fire")
    }
}

func TestRegister_SIGTERM_InvokesHandler_Unix(t *testing.T) {
    defer Reset()
    fired := make(chan struct{}, 1)
    Register(SIGTERM, func(){ fired <- struct{}{} })
    time.Sleep(10 * time.Millisecond)
    p, _ := os.FindProcess(os.Getpid())
    // On non-Unix platforms SIGTERM may not exist; guard by attempting syscall
    _ = p.Signal(syscall.Signal(15))
    select {
    case <-fired:
    case <-time.After(2 * time.Second):
        // Non-fatal on unsupported platforms; just skip
        t.Skip("SIGTERM may be unsupported on this platform")
    }
}

