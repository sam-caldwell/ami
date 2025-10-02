package ir

type Dispatch struct{ Label string }

func (d Dispatch) isInstruction() Kind { return OpDispatch }

