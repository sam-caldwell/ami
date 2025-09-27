package ir

// ExecContext is a scaffold placeholder for future runtime/execution integration.
// It captures conceptual execution settings without affecting codegen semantics yet.
type ExecContext struct {
    // Sandbox indicates whether execution is restricted (no ambient I/O, etc.).
    Sandbox bool `json:"sandbox,omitempty"`
    // Limits captures optional resource limits (counts, bytes, milliseconds).
    Limits map[string]int64 `json:"limits,omitempty"`
    // Env captures named environment toggles relevant to workers.
    Env map[string]string `json:"env,omitempty"`
}

