package amsignal

import (
    "os"
    "runtime"
    "syscall"
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

