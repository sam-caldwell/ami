package ir

// Instruction is implemented by all instruction nodes.
type Instruction interface{ isInstruction() Kind }

