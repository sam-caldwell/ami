package llvm

import (
    "os"
    "path/filepath"
    "strconv"
)

// WriteIngressEntrypointLL writes an LLVM module with a main() that calls
// ami_rt_spawn_ingress(i8* name) for each provided ingress identifier.
// The identifiers should be stable (e.g., "pkg.pipeline").
func WriteIngressEntrypointLL(dir, triple string, ingress []string) (string, error) {
    if triple == "" { triple = DefaultTriple }
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    // Build deterministic IR text
    s := "; ModuleID = \"ami:entry\"\n"
    s += "target triple = \"" + triple + "\"\n\n"
    // extern spawn symbol (implemented by runtime or linked later)
    s += "declare void @ami_rt_spawn_ingress(ptr)\n\n"
    // ensure GPU backends are probed before any ingress work
    s += "declare void @ami_rt_gpu_probe_init()\n\n"
    // Emit string constants for each ingress
    for i, name := range ingress {
        // encode as C string with \00 terminator; escape quotes minimally
        esc := encodeCString(name)
        // array length = byte length + terminator
        n := len(name) + 1
        s += "@.ingress.str." + strconv.Itoa(i) + " = private constant [" + strconv.Itoa(n) + " x i8] c\"" + esc + "\"\n"
    }
    s += "\n"
    // main body
    s += "define i32 @main() {\nentry:\n  call void @ami_rt_gpu_probe_init()\n"
    for i, name := range ingress {
        // match the constant array length
        n := len(name) + 1
        s += "  %p" + strconv.Itoa(i) + " = getelementptr inbounds [" + strconv.Itoa(n) + " x i8], ptr @.ingress.str." + strconv.Itoa(i) + ", i64 0, i64 0\n"
        s += "  call void @ami_rt_spawn_ingress(ptr %p" + strconv.Itoa(i) + ")\n"
    }
    s += "  ret i32 0\n}\n"
    out := filepath.Join(dir, "entry.ll")
    return out, os.WriteFile(out, []byte(s), 0o644)
}

// encodeCString encodes s into a C-style bytes string suitable for LLVM c"..." with \00 terminator.
// encodeCString moved to entry_encode.go to satisfy single-declaration rule
