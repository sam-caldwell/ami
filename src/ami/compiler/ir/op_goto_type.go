package ir

type Goto struct{ Label string }

func (g Goto) isInstruction() Kind { return OpGoto }

