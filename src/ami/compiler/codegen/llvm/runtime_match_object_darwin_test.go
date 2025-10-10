//go:build darwin

package llvm

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// On Darwin, validate that the runtime includes ami_rt_match_key and ami_rt_object_find
// and that the LLVM compiles to an object without linking a shared library.
func TestRuntime_ObjectKeyMatchers_CompileOnly_Darwin(t *testing.T) {
    clang, err := FindClang()
    if err != nil { t.Skip("clang not found; skipping") }
    if ver, err := Version(clang); err == nil && ver == "" { t.Skip("clang version undetectable; skipping") }

    dir := t.TempDir()
    triple := DefaultTriple
    llp, err := WriteRuntimeLL(dir, triple, false)
    if err != nil { t.Fatalf("WriteRuntimeLL: %v", err) }
    b, err := os.ReadFile(llp)
    if err != nil { t.Fatalf("read: %v", err) }
    s := string(b)
    if !strings.Contains(s, "define i1 @ami_rt_match_key") { t.Fatalf("match_key missing in runtime.ll") }
    if !strings.Contains(s, "define ptr @ami_rt_object_find") { t.Fatalf("object_find missing in runtime.ll") }
    outObj := filepath.Join(dir, "runtime.o")
    if err := CompileLLToObject(clang, llp, outObj, triple); err != nil {
        t.Fatalf("compile runtime.ll failed: %v", err)
    }
}
