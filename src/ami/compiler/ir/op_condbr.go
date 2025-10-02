package ir

// CondBr branches to True or False labels based on a boolean condition value.
type CondBr struct {
    Cond       Value
    TrueLabel  string
    FalseLabel string
}

func (c CondBr) isInstruction() Kind { return OpCondBr }

