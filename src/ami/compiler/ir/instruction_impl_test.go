package ir

// Compile-time assertions that instruction structs implement Instruction.
var _ Instruction = Var{}
var _ Instruction = Assign{}
var _ Instruction = Return{}
var _ Instruction = Defer{}
var _ Instruction = Expr{}

