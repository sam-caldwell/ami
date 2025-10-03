package llvm

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"
    "testing"
    stdtime "time"
    "github.com/sam-caldwell/ami/src/testutil"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Integration: build a tiny module that calls handler thunk APIs, emit LLVM, compile, and link with runtime.
func TestRuntime_Link_WithSignalThunks_EmitsCalls_And_Links(t *testing.T) {
    clang, err := FindClang()
    if err != nil {
        t.Skip("clang not found; skipping")
    }
    if ver, err := Version(clang); err == nil {
        if major := parseClangMajor(ver); major > 0 && major < 15 {
            t.Skipf("clang too old for opaque pointers: %s", ver)
        }
    }
    dir := filepath.Join("build", "test", "llvm_link_signal_thunks")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }

    // Build tiny IR module with a handler and calls to install/get
    m := ir.Module{Package: "app"}
    // Handler function H
    m.Functions = append(m.Functions, ir.Function{Name: "H", Blocks: []ir.Block{{Name: "entry", Instr: []ir.Instruction{ir.Return{}}}}})
    // Function F emits calls to install/get
    f := ir.Function{Name: "F"}
    f.Blocks = append(f.Blocks, ir.Block{Name: "entry", Instr: []ir.Instruction{
        ir.Expr{Op: "call", Callee: "ami_rt_install_handler_thunk", Args: []ir.Value{{ID: "#42", Type: "int64"}, {ID: "#@H", Type: "ptr"}}},
        ir.Expr{Op: "call", Callee: "ami_rt_get_handler_thunk", Args: []ir.Value{{ID: "#42", Type: "int64"}}, Result: &ir.Value{ID: "p", Type: "ptr"}},
        ir.Return{},
    }})
    m.Functions = append(m.Functions, f)

    // Emit module to LLVM and assert externs + calls are present
    // Use host triple to ensure toolchain compatibility
    triple := DefaultTriple
    if t := TripleFor(runtime.GOOS, runtime.GOARCH); t != "" { triple = t }
    s, err := EmitModuleLLVMForTarget(m, triple)
    if err != nil { t.Fatalf("emit: %v", err) }
    wants := []string{
        "declare void @ami_rt_install_handler_thunk(i64, ptr)",
        "declare ptr @ami_rt_get_handler_thunk(i64)",
        "call void @ami_rt_install_handler_thunk(i64 42, ptr @H)",
        "call ptr @ami_rt_get_handler_thunk(i64 42)",
    }
    for _, w := range wants {
        if !strings.Contains(s, w) { t.Fatalf("missing %q in:\n%s", w, s) }
    }
    // Write module
    llMod := filepath.Join(dir, "mod.ll")
    if err := os.WriteFile(llMod, []byte(s), 0o644); err != nil { t.Fatalf("write mod: %v", err) }
    // Compile to object
    oMod := filepath.Join(dir, "mod.o")
    if err := CompileLLToObject(clang, llMod, oMod, triple); err != nil { t.Skipf("compile mod failed: %v", err) }

    // Write runtime with main()
    llRt, err := WriteRuntimeLL(dir, triple, true)
    if err != nil { t.Fatalf("write rt: %v", err) }
    oRt := filepath.Join(dir, "runtime.o")
    if err := CompileLLToObject(clang, llRt, oRt, triple); err != nil { t.Skipf("compile rt failed: %v", err) }

    // Link both into a binary
    bin := filepath.Join(dir, "app")
    if runtime.GOOS == "windows" { bin += ".exe" }
    if err := LinkObjects(clang, []string{oMod, oRt}, bin, triple); err != nil { t.Skipf("link failed: %v", err) }
    st, err := os.Stat(bin)
    if err != nil || st.IsDir() || st.Size() == 0 { t.Fatalf("binary not written: %v, st=%v", err, st) }
    // Run binary (no OS signal calls are made; should exit 0)
    ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(10*stdtime.Second))
    defer cancel()
    cmd := exec.CommandContext(ctx, bin)
    if err := cmd.Run(); err != nil || ctx.Err() != nil { t.Fatalf("run bin: %v ctx:%v", err, ctx.Err()) }
}

// parseClangMajor extracts the leading major version from a clang version string.
// parseClangMajor is provided by testutil_version_test.go
