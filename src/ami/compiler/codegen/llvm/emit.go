package llvm

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "strings"
)

// moved to emit_module.go

// EmitModuleLLVMForTarget lowers an IR module to LLVM IR using a specific target triple.
func EmitModuleLLVMForTarget(m ir.Module, triple string) (string, error) {
    e := NewModuleEmitter(m.Package, "")
    if triple != "" { e.SetTargetTriple(triple) }
    // Collect externs based on usage
    for _, f := range m.Functions {
        for _, b := range f.Blocks {
            for _, ins := range b.Instr {
                switch v := ins.(type) {
                case ir.Expr:
                    addExternsForExpr(e, v)
                case ir.Defer:
                    addExternsForExpr(e, v.Expr)
                }
            }
        }
    }
    // Embed error pipeline metadata as globals for backend/runtime discovery
    for _, ep := range m.ErrorPipes {
        // Construct a simple, deterministic payload: "pipeline:<name>|steps:<s1>,<s2>,..."
        payload := "pipeline:" + ep.Pipeline + "|steps:"
        for i, s := range ep.Steps {
            if i > 0 { payload += "," }
            payload += s
        }
        // LLVM string constant: @ami_errpipe_<pipeline> = private constant [N x i8] c"...\00"
        // Assemble definition with proper length
        n := len(payload) + 1 // include NUL
        // Escape backslashes and quotes to be safe
        esc := strings.ReplaceAll(payload, "\\", "\\5C")
        esc = strings.ReplaceAll(esc, "\"", "\\22")
        name := "@ami_errpipe_" + sanitizeIdent(ep.Pipeline)
        def := name + " = private constant [" + itoa(n) + " x i8] c\"" + esc + "\\00\""
        e.AddGlobal(def)
    }
    // Emit module metadata as a single JSON string constant for runtime discovery
    if meta := buildModuleMetaJSON(m); meta != "" {
        n := len(meta) + 1
        name := "@ami_meta_json"
        // escape quotes and backslashes for c"..."
        esc := strings.ReplaceAll(meta, "\\", "\\5C")
        esc = strings.ReplaceAll(esc, "\"", "\\22")
        def := name + " = private constant [" + itoa(n) + " x i8] c\"" + esc + "\\00\""
        e.AddGlobal(def)
    }
    for _, f := range m.Functions {
        if err := e.AddFunction(f); err != nil { return "", err }
    }
    return e.Build(), nil
}

// addExternsForExpr moved to add_externs_for_expr.go to satisfy single-declaration rule
