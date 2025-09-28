package llvm

import (
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "testing"
)

func TestRuntime_CompileAndLink_ProducesExecutable(t *testing.T) {
    clang, err := FindClang()
    if err != nil {
        t.Skip("clang not found; skipping runtime link test")
    }
    dir := filepath.Join("build", "test", "llvm_runtime_link")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Write runtime with main()
    ll, err := WriteRuntimeLL(dir, DefaultTriple, true)
    if err != nil { t.Fatalf("write ll: %v", err) }
    // Compile to object
    obj := filepath.Join(dir, "runtime.o")
    if err := CompileLLToObject(clang, ll, obj, DefaultTriple); err != nil { t.Fatalf("compile ll -> o: %v", err) }
    // Link into binary
    bin := filepath.Join(dir, "app")
    if runtime.GOOS == "windows" { bin += ".exe" }
    if err := LinkObjects(clang, []string{obj}, bin, DefaultTriple); err != nil { t.Fatalf("link: %v", err) }
    st, err := os.Stat(bin)
    if err != nil || st.IsDir() || st.Size() == 0 { t.Fatalf("binary not written: %v, st=%v", err, st) }
    // Execute and expect exit code 0
    cmd := exec.Command(bin)
    if err := cmd.Run(); err != nil { t.Fatalf("run binary: %v", err) }
}
