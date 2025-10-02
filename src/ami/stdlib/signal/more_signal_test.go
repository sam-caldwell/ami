package amsignal

import (
    "os"
    "syscall"
    "testing"
)

func TestFromOSSignal_Mapping(t *testing.T) {
    if fromOSSignal(os.Interrupt) != SIGINT { t.Fatalf("Interrupt -> SIGINT") }
    // These may not be present on all platforms; best-effort checks
    if fromOSSignal(syscall.SIGTERM) != SIGTERM { t.Fatalf("SIGTERM mapping") }
    _ = fromOSSignal(syscall.SIGHUP)
    _ = fromOSSignal(syscall.SIGQUIT)
}

func TestSafeCall_HandlesNilAndPanic(t *testing.T) {
    // nil should be ignored
    safeCall(nil)
    // panicking function should not crash
    safeCall(func(){ panic("boom") })
}

func TestToOSSignal_AllCases(t *testing.T) {
    // Ensure mapping returns a non-nil os.Signal for each known type
    if toOSSignal(SIGINT) == nil { t.Fatalf("SIGINT toOSSignal nil") }
    if toOSSignal(SIGTERM) == nil { t.Fatalf("SIGTERM toOSSignal nil") }
    if toOSSignal(SIGHUP) == nil { t.Fatalf("SIGHUP toOSSignal nil") }
    if toOSSignal(SIGQUIT) == nil { t.Fatalf("SIGQUIT toOSSignal nil") }
}

