package driver

// Options controls compilation behavior.
type Options struct {
    // Debug enables emission of debug artifacts under build/debug.
    Debug bool
    // EmitLLVMOnly emits .ll but skips compiling to .o
    EmitLLVMOnly bool
    // NoLink skips linking (reserved; backend pending)
    NoLink bool
    // Log, if non-nil, receives activity events for timestamped logging by caller.
    // The event string should be a short dot-delimited identifier, and fields
    // is an optional map with structured context. The logger is expected to
    // apply timestamps.
    Log func(event string, fields map[string]any)
}
