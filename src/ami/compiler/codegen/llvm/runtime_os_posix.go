//go:build runtime_posix && (darwin || linux || freebsd)

package llvm

// runtimeOSLL appends POSIX-oriented trampolines that demonstrate how OS signal
// numbers map to AMI handler tokens and dispatch via handler thunks. This avoids
// actual syscalls; installation with sigaction will be added when hooks are approved.
func runtimeOSLL() string {
    s := "; POSIX signal trampolines (build-tag: runtime_posix)\n"
    // Helper: return current handler token for a given OS signal number.
    // Mapping: idx = signum % 64, then read @ami_signal_handlers[idx].
    s += "define i64 @ami_rt_signal_token_for(i64 %sig) {\n" +
        "entry:\n  %idx = urem i64 %sig, 64\n  %slot = getelementptr [64 x i64], ptr @ami_signal_handlers, i64 0, i64 %idx\n  %tok = load i64, ptr %slot, align 8\n  ret i64 %tok\n}\n\n"
    // Trampoline: called by OS layer for a signal (signum: i32). Fetch token, get thunk, and call if present.
    s += "define void @ami_rt_posix_trampoline(i32 %signum) {\n" +
        "entry:\n  %s64 = zext i32 %signum to i64\n  %tok = call i64 @ami_rt_signal_token_for(i64 %s64)\n  %fp = call ptr @ami_rt_get_handler_thunk(i64 %tok)\n  %isnull = icmp eq ptr %fp, null\n  br i1 %isnull, label %exit, label %call\n" +
        "call:\n  %cast = bitcast ptr %fp to void ()*\n  call void %cast()\n  br label %exit\n" +
        "exit:\n  ret void\n}\n\n"
    // Stubs for enable/disable (no-ops now; wire to sigaction later)
    s += "define void @ami_rt_os_signal_enable(i64 %sig) {\nentry:\n  ret void\n}\n\n"
    s += "define void @ami_rt_os_signal_disable(i64 %sig) {\nentry:\n  ret void\n}\n\n"
    return s
}
