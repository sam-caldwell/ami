package llvm

import (
    "fmt"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "strings"
)

// EmitWorkerWrappersOnlyForTarget emits only the ami_worker_core_<name> wrappers for worker-shaped
// functions and declares the worker symbols without defining them. This allows linking external
// definitions of the worker bodies (e.g., in C) without duplicate symbol conflicts.
func EmitWorkerWrappersOnlyForTarget(m ir.Module, triple string) (string, error) {
    e := NewModuleEmitter(m.Package, "")
    if triple != "" { e.SetTargetTriple(triple) }
    // Declare worker symbols with ABI-safe signatures: {ty0,ty1} @Name(i64)
    for _, fn := range m.Functions {
        if len(fn.Params) != 1 || len(fn.Results) != 2 { continue }
        if !strings.HasPrefix(fn.Params[0].Type, "Event<") || fn.Results[1].Type != "error" { continue }
        ty0 := abiType(fn.Results[0].Type)
        ty1 := abiType(fn.Results[1].Type)
        decl := fmt.Sprintf("declare { %s, %s } @%s(i64)", ty0, ty1, fn.Name)
        e.RequireExtern(decl)
    }
    // Emit wrappers that call declared worker symbols
    emitWorkerCores(e, m.Functions)
    return e.Build(), nil
}

