package ir

// Expr is a generic expression instruction (call/op) emitting a value.
type Expr struct {
    Op     string   // operation name or callee
    Args   []Value  // arguments
    Result *Value   // optional result
}

func (e Expr) isInstruction() Kind { return OpExpr }

