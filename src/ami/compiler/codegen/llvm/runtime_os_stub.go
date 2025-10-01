//go:build !runtime_posix

package llvm

// runtimeOSLL returns OS-specific LLVM IR snippets to be appended to the runtime
// when enabled via build tags. Default (no tags): provide stub enable/disable.
func runtimeOSLL() string {
    s := "define void @ami_rt_os_signal_enable(i64 %sig) {\nentry:\n  ret void\n}\n\n"
    s += "define void @ami_rt_os_signal_disable(i64 %sig) {\nentry:\n  ret void\n}\n\n"
    return s
}
