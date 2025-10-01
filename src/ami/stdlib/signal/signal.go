package amsignal

import (
    "os"
    gosignal "os/signal"
    "runtime"
    "sync"
    "syscall"
)

// SignalType is an enum of OS signals supported by Register.
type SignalType int

const (
    SIGINT SignalType = iota + 1
    SIGTERM
    SIGHUP
    SIGQUIT
)

// toOSSignal maps SignalType to a concrete os.Signal for the current platform.
func toOSSignal(s SignalType) os.Signal {
    switch s {
    case SIGINT:
        // On Windows, Interrupt is the closest
        if runtime.GOOS == "windows" { return os.Interrupt }
        return syscall.SIGINT
    case SIGTERM:
        if runtime.GOOS == "windows" { return os.Kill }
        return syscall.SIGTERM
    case SIGHUP:
        if runtime.GOOS == "windows" { return os.Kill }
        return syscall.SIGHUP
    case SIGQUIT:
        if runtime.GOOS == "windows" { return os.Kill }
        return syscall.SIGQUIT
    default:
        if runtime.GOOS == "windows" { return os.Kill }
        return syscall.SIGTERM
    }
}

var (
    mu       sync.Mutex
    once     sync.Once
    ch       chan os.Signal
    handlers = map[SignalType][]func(){}
)

// Register installs fn as a handler for the given signal. Multiple handlers may be
// registered per signal. Handlers are invoked sequentially when a signal arrives.
func Register(sig SignalType, fn func()) {
    mu.Lock()
    defer mu.Unlock()
    handlers[sig] = append(handlers[sig], fn)
    // ensure goroutine started once; subscribe for signals we know about
    once.Do(start)
    gosignal.Notify(ch, toOSSignal(sig))
}

// start initializes the shared signal channel and dispatcher loop.
func start() {
    ch = make(chan os.Signal, 4)
    go func() {
        for s := range ch {
            // Map os.Signal back to our enum set
            st := fromOSSignal(s)
            if st == 0 { continue }
            mu.Lock()
            fns := append([]func(){}, handlers[st]...)
            mu.Unlock()
            for _, f := range fns { safeCall(f) }
        }
    }()
}

func safeCall(f func()) {
    defer func(){ _ = recover() }()
    if f != nil { f() }
}

// fromOSSignal best-effort conversion from os.Signal to SignalType for our set.
func fromOSSignal(s os.Signal) SignalType {
    switch s {
    case os.Interrupt, syscall.SIGINT:
        return SIGINT
    case syscall.SIGTERM:
        return SIGTERM
    case syscall.SIGHUP:
        return SIGHUP
    case syscall.SIGQUIT:
        return SIGQUIT
    }
    return 0
}

// Reset clears all registered handlers and stops notifications (for tests).
func Reset() {
    mu.Lock()
    defer mu.Unlock()
    for st := range handlers { handlers[st] = nil }
    if ch != nil {
        gosignal.Stop(ch)
        close(ch)
        ch = nil
    }
    // allow re-init
    once = sync.Once{}
}

