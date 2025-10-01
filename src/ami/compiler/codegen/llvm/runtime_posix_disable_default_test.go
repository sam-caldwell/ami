//go:build runtime_posix && (darwin || linux || freebsd)

package llvm

import (
    "strings"
    "testing"
)

// Validate that Disable restores default handler via signal(sig, SIG_DFL) in IR.
func TestRuntime_Posix_Disable_Restores_Default(t *testing.T) {
    s := RuntimeLL(DefaultTriple, false)
    if !strings.Contains(s, "define void @ami_rt_os_signal_disable(i64 %sig)") {
        t.Fatalf("missing disable function in IR:\n%s", s)
    }
    if !strings.Contains(s, "call ptr @signal(i32 %s32, ptr null)") {
        t.Fatalf("disable does not call signal with null handler:\n%s", s)
    }
}

