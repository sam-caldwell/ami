package llvm

import (
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "testing"
)

// Verify WriteIngressEntrypointLL emits a main() calling the spawn helper for each ingress name
// and encodes string constants with correct lengths.
func TestWriteIngressEntrypointLL_EmitsMainAndSpawnCalls(t *testing.T) {
    dir := filepath.Join("build", "test", "entry_write")
    _ = os.RemoveAll(dir)
    names := []string{"pkg1.ingressA", "pkg2.ingressB"}
    path, err := WriteIngressEntrypointLL(dir, DefaultTriple, names)
    if err != nil { t.Fatalf("write entry: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read entry: %v", err) }
    s := string(b)
    if !strings.Contains(s, "ModuleID = \"ami:entry\"") {
        t.Fatalf("missing entry header: %s", s)
    }
    if !strings.Contains(s, "declare void @ami_rt_spawn_ingress(ptr)") {
        t.Fatalf("missing spawn extern: %s", s)
    }
    // Ensure each name has a corresponding constant and call
    for i, n := range names {
        wantConst := "@.ingress.str." + strconv.Itoa(i)
        if !strings.Contains(s, wantConst) { t.Fatalf("missing const %q in %s", wantConst, s) }
        // length should be len(name)+1 in the array type
        ln := len(n) + 1
        if !strings.Contains(s, "["+strconv.Itoa(ln)+" x i8]") {
            t.Fatalf("missing array size for %q: %s", n, s)
        }
        if !strings.Contains(s, "call void @ami_rt_spawn_ingress(ptr %p"+strconv.Itoa(i)+")") {
            t.Fatalf("missing spawn call for #%d in %s", i, s)
        }
    }
}
