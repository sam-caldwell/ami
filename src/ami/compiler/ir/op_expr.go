package ir

// Expr is a generic expression instruction (call/op) emitting a value.
type Expr struct {
    Op     string   // operation kind (e.g., call, add)
    Callee string   // callee name when Op=="call"
    Args   []Value  // arguments
    Result *Value   // optional result
    // Results holds multiple results for tuple-returning expressions (e.g., multi-value calls).
    // When non-empty, Result is ignored.
    Results []Value
    // Debug signature info for calls (JSON-only debug emission)
    ParamTypes  []string
    ParamNames  []string
    ResultTypes []string
}

func (e Expr) isInstruction() Kind { return OpExpr }
