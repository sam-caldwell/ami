//go:build runtime_posix && (darwin || linux || freebsd)

package llvm

// runtimeOSLL appends POSIX-oriented scaffolding for future signal integration.
// This intentionally avoids real syscalls until hooks are approved.
func runtimeOSLL() string {
    // Placeholders for OS-level install/shim. No-ops for now.
    // When enabled, these can reference libc signal APIs and dispatch into handler thunks.
    s := "; POSIX signal scaffolding (build-tag: runtime_posix)\n"
    s += "define void @ami_rt_os_signal_enable(i64 %sig) {\nentry:\n  ret void\n}\n\n"
    s += "define void @ami_rt_os_signal_disable(i64 %sig) {\nentry:\n  ret void\n}\n\n"
    return s
}

