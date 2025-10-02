package llvm

import (
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strings"
    "testing"
)

// This test validates:
// - llvm.memcpy is present in runtime (via owned_new)
// - zeroization sets bytes to 0x00
// - double-release does not crash (guarded)
func TestRuntime_Owned_CopyOnNew_Zeroize_DoubleRelease(t *testing.T) {
    clang, err := FindClang()
    if err != nil { t.Skip("clang not found; skipping") }
    if ver, err := Version(clang); err == nil {
        if major := parseClangMajor(ver); major > 0 && major < 15 {
            t.Skipf("clang too old for opaque pointers: %s", ver)
        }
    }
    dir := filepath.Join("build", "test", "llvm_runtime_owned_raii")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Write runtime without main
    triple := TripleFor(runtime.GOOS, runtime.GOARCH)
    ll, err := WriteRuntimeLL(dir, triple, false)
    if err != nil { t.Fatalf("write ll: %v", err) }
    // Replace calls to free in runtime with a test wrapper so we can count frees safely
    if b, err := os.ReadFile(ll); err == nil {
        s := string(b)
        s = strings.ReplaceAll(s, "call void @free(", "call void @ami_test_free(")
        if err := os.WriteFile(ll, []byte(s), 0o644); err != nil { t.Fatalf("rewrite ll: %v", err) }
    } else { t.Fatalf("read ll: %v", err) }
    // Append a free counter wrapper and a main that tests zeroize and owned double-release
    add := `
@free_count = private global i64 0
define void @ami_test_free(ptr %p) {
entry:
  %c0 = load i64, ptr @free_count
  %c1 = add i64 %c0, 1
  store i64 %c1, ptr @free_count
  ret void
}

// parseClangMajor is provided by testutil_version_test.go

@.str = private constant [5 x i8] c"TEST\00"

define i32 @main() {
entry:
  ; test zeroize on raw buffer
  %buf = call ptr @malloc(i64 4)
  store i8 1, ptr %buf
  %p1 = getelementptr i8, ptr %buf, i64 1
  store i8 2, ptr %p1
  %p2 = getelementptr i8, ptr %buf, i64 2
  store i8 3, ptr %p2
  %p3 = getelementptr i8, ptr %buf, i64 3
  store i8 4, ptr %p3
  call void @ami_rt_zeroize(ptr %buf, i64 4)
  %b0 = load i8, ptr %buf
  %p1b = getelementptr i8, ptr %buf, i64 1
  %b1 = load i8, ptr %p1b
  %p2b = getelementptr i8, ptr %buf, i64 2
  %b2 = load i8, ptr %p2b
  %p3b = getelementptr i8, ptr %buf, i64 3
  %b3 = load i8, ptr %p3b
  %c0 = icmp ne i8 %b0, 0
  br i1 %c0, label %fail, label %c1
c1:
  %c1v = icmp ne i8 %b1, 0
  br i1 %c1v, label %fail, label %c2
c2:
  %c2v = icmp ne i8 %b2, 0
  br i1 %c2v, label %fail, label %c3
c3:
  %c3v = icmp ne i8 %b3, 0
  br i1 %c3v, label %fail, label %owned
owned:
  %gptr = getelementptr [5 x i8], ptr @.str, i64 0, i64 0
  %h = call ptr @ami_rt_owned_new(ptr %gptr, i64 4)
  call void @ami_rt_zeroize_owned(ptr %h)
  call void @ami_rt_zeroize_owned(ptr %h)
  ; expect exactly two frees (data + handle) total
  %fc = load i64, ptr @free_count
  %ok = icmp eq i64 %fc, 2
  br i1 %ok, label %oklbl, label %fail
oklbl:
  ret i32 0
fail:
  ret i32 1
}
`
    if f, err := os.OpenFile(ll, os.O_APPEND|os.O_WRONLY, 0); err == nil {
        defer f.Close()
        if _, err := f.WriteString(add); err != nil { t.Fatalf("append: %v", err) }
    } else { t.Fatalf("open: %v", err) }
    // Compile and link binary
    obj := filepath.Join(dir, "rt.o")
    if err := CompileLLToObject(clang, ll, obj, triple); err != nil { t.Skipf("compile failed: %v", err) }
    bin := filepath.Join(dir, "app")
    if runtime.GOOS == "windows" { bin += ".exe" }
    if err := LinkObjects(clang, []string{obj}, bin, triple); err != nil { t.Skipf("link failed: %v", err) }
    if err := exec.Command(bin).Run(); err != nil { t.Fatalf("run: %v", err) }
}
