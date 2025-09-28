package llvm

// ToolError wraps a tool invocation failure with captured stderr for diagnostics.
type ToolError struct {
    Tool   string
    Stderr string
}

func (e ToolError) Error() string { return e.Tool + " failed" }

