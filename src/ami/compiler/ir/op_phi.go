package ir

// Phi selects a value based on the predecessor block.
type Phi struct {
    Result    Value
    Incomings []PhiIncoming
}

func (p Phi) isInstruction() Kind { return OpPhi }
