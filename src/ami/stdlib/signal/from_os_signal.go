package amsignal

import (
    "os"
    "syscall"
)

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

