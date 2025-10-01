package llvm

// runtimeOSLL returns OS-specific LLVM IR snippets to be appended to the runtime
// when enabled via build tags. Default (no tags): empty.
func runtimeOSLL() string { return "" }

