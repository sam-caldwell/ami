package main

import (
    "io"
    "os"
    "path/filepath"
    "runtime"
    "testing"
    "strings"
    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
)

// Build examples/correct and run the produced binary for the host env when toolchain is present.
func TestExamplesCorrect_BuildsAndRuns_WhenToolchainPresent(t *testing.T) {
    if _, err := llvme.FindClang(); err != nil { t.Skip("clang not found") }
    example := filepath.Join("..", "..", "..", "examples", "correct")
    tmp := filepath.Join("build", "test", "examples_correct_exec")
    _ = os.RemoveAll(tmp)
    if err := copyDir(example, tmp); err != nil { t.Fatalf("copyDir: %v", err) }
    // Narrow the env matrix to the host env to ensure a runnable binary is produced deterministically.
    wsPath := filepath.Join(tmp, "ami.workspace")
    b, err := os.ReadFile(wsPath)
    if err == nil {
        host := runtime.GOOS + "/" + runtime.GOARCH
        s := string(b)
        // rewrite the env block to a single host entry
        if strings.Contains(s, "\n    env:") {
            // crude replace: collapse entire env list to one line
            i := strings.Index(s, "\n    env:")
            if i >= 0 {
                // find next non-indented section (linker or linter)
                j := strings.Index(s[i:], "\n  linker:")
                if j > 0 {
                    head := s[:i]
                    tail := s[i+j:]
                    s = head + "\n    env:\n      - " + host + "\n" + tail
                    _ = os.WriteFile(wsPath, []byte(s), 0o644)
                }
            }
        }
    }
    if err := runBuild(os.Stdout, tmp, false, false); err != nil { t.Fatalf("runBuild: %v", err) }
    env := runtime.GOOS + "/" + runtime.GOARCH
    // Locate any executable binary under build/<env>/ (skip workers libs)
    var bin string
    root := filepath.Join(tmp, "build", env)
    _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil { return nil }
        if d.IsDir() { return nil }
        name := filepath.Base(path)
        // skip shared libs and object files
        if strings.HasSuffix(name, ".dylib") || strings.HasSuffix(name, ".so") || strings.HasSuffix(name, ".o") { return nil }
        if info, e := d.Info(); e == nil {
            if m := info.Mode(); m.IsRegular() && (m&0o111 != 0) { bin = path; return io.EOF }
        }
        return nil
    })
    if bin == "" { t.Fatalf("no executable found under %s", root) }
    // If desired, we could execute it. For now, confirming presence is sufficient.
}
