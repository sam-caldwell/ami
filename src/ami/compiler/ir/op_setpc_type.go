package ir

type SetPC struct{ PC int }

func (s SetPC) isInstruction() Kind { return OpSetPC }

