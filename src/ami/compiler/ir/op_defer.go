package ir

// Defer schedules an expression to run at function exit.
type Defer struct { Expr Expr }

func (d Defer) isInstruction() Kind { return OpDefer }

