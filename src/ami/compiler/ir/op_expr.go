package ir

// Expr is a generic expression instruction (call/op) emitting a value.
type Expr struct {
    Op     string   // operation kind (e.g., call, add)
    Callee string   // callee name when Op=="call"
    Args   []Value  // arguments
    Result *Value   // optional result
}

func (e Expr) isInstruction() Kind { return OpExpr }
