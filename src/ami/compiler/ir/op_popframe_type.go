package ir

type PopFrame struct{}

func (p PopFrame) isInstruction() Kind { return OpPopFrame }

