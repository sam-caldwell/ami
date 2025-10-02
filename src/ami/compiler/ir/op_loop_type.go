package ir

type Loop struct{ Name string }

func (l Loop) isInstruction() Kind { return OpLoop }

