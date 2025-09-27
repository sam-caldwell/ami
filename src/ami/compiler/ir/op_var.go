package ir

// Var declares a new variable with optional initializer.
type Var struct {
    Name   string
    Type   string
    Init   *Value // optional
    Result Value  // variable value id
}

func (v Var) isInstruction() Kind { return OpVar }

