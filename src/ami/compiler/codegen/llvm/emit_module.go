package llvm

import "github.com/sam-caldwell/ami/src/ami/compiler/ir"

// EmitModuleLLVM lowers an IR module using the default target triple.
func EmitModuleLLVM(m ir.Module) (string, error) {
    return EmitModuleLLVMForTarget(m, "")
}

