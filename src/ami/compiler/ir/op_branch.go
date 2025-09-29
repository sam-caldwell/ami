package ir

// CondBr branches to True or False labels based on a boolean condition value.
type CondBr struct {
    Cond      Value
    TrueLabel string
    FalseLabel string
}
func (c CondBr) isInstruction() Kind { return OpCondBr }

// Phi selects a value based on the predecessor block.
type Phi struct {
    Result   Value
    Incomings []PhiIncoming
}
func (p Phi) isInstruction() Kind { return OpPhi }

type PhiIncoming struct{
    Value Value
    Label string
}

