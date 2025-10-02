package ir

type PushFrame struct{ Fn string }

func (p PushFrame) isInstruction() Kind { return OpPushFrame }

