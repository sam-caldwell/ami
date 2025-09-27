package ir

// Assign copies source into destination variable id.
type Assign struct {
    DestID string
    Src    Value
}

func (a Assign) isInstruction() Kind { return OpAssign }

