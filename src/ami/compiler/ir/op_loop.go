package ir

type Loop struct{ Name string }
func (l Loop) isInstruction() Kind { return OpLoop }

type Goto struct{ Label string }
func (g Goto) isInstruction() Kind { return OpGoto }

type SetPC struct{ PC int }
func (s SetPC) isInstruction() Kind { return OpSetPC }

type Dispatch struct{ Label string }
func (d Dispatch) isInstruction() Kind { return OpDispatch }

type PushFrame struct{ Fn string }
func (p PushFrame) isInstruction() Kind { return OpPushFrame }

type PopFrame struct{}
func (p PopFrame) isInstruction() Kind { return OpPopFrame }

