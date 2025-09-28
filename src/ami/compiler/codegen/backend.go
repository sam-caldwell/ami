package codegen

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Backend abstracts code generation and linking for a target toolchain.
// Implementations may use different codegen strategies (e.g., LLVM, custom backends).
type Backend interface {
    // Name returns a short identifier like "llvm".
    Name() string
    // EmitModule returns textual IR (backend-specific) for the given module.
    EmitModule(m ir.Module) (string, error)
    // EmitModuleForTarget emits textual IR for a specific target triple.
    EmitModuleForTarget(m ir.Module, triple string) (string, error)
    // CompileLLToObject compiles textual IR at llPath into an object file for the target triple.
    CompileLLToObject(clang, llPath, outObj, targetTriple string) error
    // LinkObjects links object files into an executable for the target triple with optional flags.
    LinkObjects(clang string, objs []string, outBin, targetTriple string, extra ...string) error
    // WriteRuntimeLL writes a minimal runtime (and optional main) for the target triple and returns its .ll path.
    WriteRuntimeLL(dir, triple string, withMain bool) (string, error)
    // FindToolchain returns the compiler path needed for compile/link steps.
    FindToolchain() (string, error)
    // ToolVersion returns a version string for the compiler (for logging/debug).
    ToolVersion(path string) (string, error)
    // TripleForEnv maps an env string like "os/arch" to a target triple.
    TripleForEnv(env string) string
}

// defaultBackend is the process-wide default backend. Initialized to LLVM.
var defaultBackend Backend = &llvmBackend{}

// DefaultBackend returns the configured default backend.
func DefaultBackend() Backend { return defaultBackend }

