package llvm

import (
    "strings"
    "testing"
)

// Ensure runtime wires Register to OS enable even without posix tag (stub present).
func TestRuntime_SignalRegister_Calls_OSEnable(t *testing.T) {
    s := RuntimeLL(DefaultTriple, false)
    if !strings.Contains(s, "define void @ami_rt_signal_register(i64 %sig, i64 %handler)") {
        t.Fatalf("missing signal_register definition:\n%s", s)
    }
    if !strings.Contains(s, "call void @ami_rt_os_signal_enable(i64 %sig)") {
        t.Fatalf("signal_register missing call to os_signal_enable:\n%s", s)
    }
}

