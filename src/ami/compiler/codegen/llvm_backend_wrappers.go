package codegen

import (
    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// EmitWorkerWrappersOnlyForTarget is a convenience entrypoint to emit only worker core wrappers
// for a given IR module and target triple. Useful for linking external worker bodies without
// duplicate symbol definitions.
func EmitWorkerWrappersOnlyForTarget(m ir.Module, triple string) (string, error) {
    return llvme.EmitWorkerWrappersOnlyForTarget(m, triple)
}

