package ir

// Return returns zero or more values from the current function.
type Return struct { Values []Value }

func (r Return) isInstruction() Kind { return OpReturn }

