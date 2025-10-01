//go:build runtime_posix && (darwin || linux || freebsd)

package llvm

import (
    "strings"
    "testing"
)

// Validate that the POSIX trampoline IR is included when built with runtime_posix.
func TestRuntime_Posix_Trampoline_IR_Present(t *testing.T) {
    s := RuntimeLL(DefaultTriple, false)
    wants := []string{
        "define i64 @ami_rt_signal_token_for(i64 %sig)",
        "define void @ami_rt_posix_trampoline(i32 %signum)",
        "declare ptr @signal(i32, ptr)",
        "define void @ami_rt_posix_install_trampoline(i64 %sig)",
        "define void @ami_rt_os_signal_enable(i64 %sig)",
        "define void @ami_rt_os_signal_disable(i64 %sig)",
    }
    for _, w := range wants {
        if !strings.Contains(s, w) {
            t.Fatalf("missing IR fragment: %q\n%s", w, s)
        }
    }
}
