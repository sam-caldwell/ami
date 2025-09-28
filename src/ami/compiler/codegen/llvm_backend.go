package codegen

import (
    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// llvmBackend adapts the llvm codegen package to the Backend interface.
type llvmBackend struct{}

func (*llvmBackend) Name() string { return "llvm" }

func (*llvmBackend) EmitModule(m ir.Module) (string, error) { return llvme.EmitModuleLLVM(m) }

func (*llvmBackend) EmitModuleForTarget(m ir.Module, triple string) (string, error) {
    return llvme.EmitModuleLLVMForTarget(m, triple)
}

func (*llvmBackend) CompileLLToObject(clang, llPath, outObj, targetTriple string) error {
    return llvme.CompileLLToObject(clang, llPath, outObj, targetTriple)
}

func (*llvmBackend) LinkObjects(clang string, objs []string, outBin, targetTriple string, extra ...string) error {
    return llvme.LinkObjects(clang, objs, outBin, targetTriple, extra...)
}

func (*llvmBackend) WriteRuntimeLL(dir, triple string, withMain bool) (string, error) {
    return llvme.WriteRuntimeLL(dir, triple, withMain)
}

func (*llvmBackend) FindToolchain() (string, error) { return llvme.FindClang() }

func (*llvmBackend) ToolVersion(path string) (string, error) { return llvme.Version(path) }

func (*llvmBackend) TripleForEnv(env string) string { return llvme.TripleForEnv(env) }

