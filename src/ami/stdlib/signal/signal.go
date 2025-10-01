package signal

import (
    ossignal "os/signal"
    "os"
    "sync"
    "syscall"
    "errors"
)

// SignalType is an enum of supported OS signals.
type SignalType string

const (
    SIGHUP  SignalType = "SIGHUP"
    SIGINT  SignalType = "SIGINT"
    SIGQUIT SignalType = "SIGQUIT"
    SIGTERM SignalType = "SIGTERM"
    SIGUSR1 SignalType = "SIGUSR1"
    SIGUSR2 SignalType = "SIGUSR2"
    SIGALRM SignalType = "SIGALRM"
    SIGCHLD SignalType = "SIGCHLD"
    SIGPIPE SignalType = "SIGPIPE"
    // Note: SIGKILL and SIGSTOP cannot be caught or handled; Register returns error for them.
    SIGKILL SignalType = "SIGKILL"
    SIGSTOP SignalType = "SIGSTOP"
)

var (
    mu       sync.Mutex
    started  bool
    ch       chan os.Signal
    // callbacks mapped by SignalType
    cbs      = map[SignalType][]func(){}
)

// Register registers a handler function to be invoked when the given signal is received.
// Returns an error if the signal is not supported (e.g., SIGKILL, SIGSTOP).
func Register(sig SignalType, fn func()) error {
    osSig, ok := toOSSignal(sig)
    if !ok { return errors.New("unsupported signal: " + string(sig)) }
    mu.Lock()
    defer mu.Unlock()
    if !started {
        ch = make(chan os.Signal, 8)
        go dispatch()
        started = true
    }
    // subscribe the channel to this signal
    ossignal.Notify(ch, osSig)
    cbs[sig] = append(cbs[sig], fn)
    return nil
}

func dispatch() {
    for s := range ch {
        sig := fromOSSignal(s)
        mu.Lock()
        handlers := append([]func(){}, cbs[sig]...)
        mu.Unlock()
        for _, h := range handlers { if h != nil { h() } }
    }
}

// toOSSignal maps a SignalType to the platform os.Signal. Returns false if unsupported.
func toOSSignal(s SignalType) (os.Signal, bool) {
    switch s {
    case SIGHUP:
        return syscall.SIGHUP, true
    case SIGINT:
        return syscall.SIGINT, true
    case SIGQUIT:
        return syscall.SIGQUIT, true
    case SIGTERM:
        return syscall.SIGTERM, true
    case SIGUSR1:
        return syscall.SIGUSR1, true
    case SIGUSR2:
        return syscall.SIGUSR2, true
    case SIGALRM:
        return syscall.SIGALRM, true
    case SIGCHLD:
        return syscall.SIGCHLD, true
    case SIGPIPE:
        return syscall.SIGPIPE, true
    case SIGKILL, SIGSTOP:
        return nil, false
    default:
        return nil, false
    }
}

// fromOSSignal converts an os.Signal back into a SignalType where possible.
func fromOSSignal(s os.Signal) SignalType {
    switch s {
    case syscall.SIGHUP:
        return SIGHUP
    case syscall.SIGINT:
        return SIGINT
    case syscall.SIGQUIT:
        return SIGQUIT
    case syscall.SIGTERM:
        return SIGTERM
    case syscall.SIGUSR1:
        return SIGUSR1
    case syscall.SIGUSR2:
        return SIGUSR2
    case syscall.SIGALRM:
        return SIGALRM
    case syscall.SIGCHLD:
        return SIGCHLD
    case syscall.SIGPIPE:
        return SIGPIPE
    default:
        return ""
    }
}

