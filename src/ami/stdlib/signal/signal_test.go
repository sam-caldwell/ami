package signal

import (
    "os"
    "runtime"
    "syscall"
    "testing"
    stdsignal "os/signal"
    "time"
)

func TestRegister_SIGUSR1_InvokesHandler(t *testing.T) {
    if runtime.GOOS == "windows" { t.Skip("signals not fully supported on windows") }
    fired := make(chan struct{}, 1)
    if err := Register(SIGUSR1, func(){ select{ case fired<-struct{}{}: default: } }); err != nil {
        t.Fatalf("Register: %v", err)
    }
    // send signal to current process
    p, err := os.FindProcess(os.Getpid())
    if err != nil { t.Fatalf("FindProcess: %v", err) }
    if err := p.Signal(syscall.SIGUSR1); err != nil { t.Fatalf("Signal: %v", err) }
    select {
    case <-fired:
        // ok
    case <-time.After(1 * time.Second):
        t.Fatalf("timeout waiting for handler")
    }
    // reset signal notifications to avoid leaking into other tests
    stdsignal.Reset()
}

func TestRegister_UnsupportedSignals_Error(t *testing.T) {
    if err := Register(SIGKILL, func(){}); err == nil {
        t.Fatalf("expected error for SIGKILL")
    }
    if err := Register(SIGSTOP, func(){}); err == nil {
        t.Fatalf("expected error for SIGSTOP")
    }
}

func TestRegister_MultipleHandlers_AllInvoked(t *testing.T) {
    if runtime.GOOS == "windows" { t.Skip("signals not fully supported on windows") }
    c1 := make(chan struct{}, 1)
    c2 := make(chan struct{}, 1)
    _ = Register(SIGUSR2, func(){ select{ case c1<-struct{}{}: default: } })
    _ = Register(SIGUSR2, func(){ select{ case c2<-struct{}{}: default: } })
    p, _ := os.FindProcess(os.Getpid())
    _ = p.Signal(syscall.SIGUSR2)
    deadline := time.After(1 * time.Second)
    var a, b bool
    for !(a && b) {
        select {
        case <-c1: a = true
        case <-c2: b = true
        case <-deadline: t.Fatalf("handlers not both invoked: a=%v b=%v", a, b)
        }
    }
    stdsignal.Reset()
}

