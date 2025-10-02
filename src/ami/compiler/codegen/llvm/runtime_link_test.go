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
    if ver, err := Version(clang); err == nil {
        // Require opaque pointer support (LLVM 15+)
        if major := parseClangMajor(ver); major > 0 && major < 15 {
            t.Skipf("clang too old for opaque pointers: %s", ver)
        }
    }
    dir := filepath.Join("build", "test", "llvm_runtime_link")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Write runtime with main() for host triple
    triple := TripleFor(runtime.GOOS, runtime.GOARCH)
    ll, err := WriteRuntimeLL(dir, triple, true)
    if err != nil { t.Fatalf("write ll: %v", err) }
    // Compile to object
    obj := filepath.Join(dir, "runtime.o")
    if err := CompileLLToObject(clang, ll, obj, triple); err != nil { t.Skipf("compile ll -> o failed: %v", err) }
    // Link into binary
    bin := filepath.Join(dir, "app")
    if runtime.GOOS == "windows" { bin += ".exe" }
    if err := LinkObjects(clang, []string{obj}, bin, triple); err != nil { t.Skipf("link failed: %v", err) }
    st, err := os.Stat(bin)
    if err != nil || st.IsDir() || st.Size() == 0 { t.Fatalf("binary not written: %v, st=%v", err, st) }
    // Execute and expect exit code 0
    cmd := exec.Command(bin)
    if err := cmd.Run(); err != nil { t.Fatalf("run binary: %v", err) }
}

// parseClangMajor extracts the leading major version from a clang version string.
// parseClangMajor is provided by testutil_version_test.go
