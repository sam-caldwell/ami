package e2e

import (
    "bytes"
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

func TestE2E_AmiBuild_Verbose_EmitsLLVM(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "build", "verbose_llvm")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Minimal workspace
    yaml := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n")
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), yaml, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // Minimal source unit
    if err := os.WriteFile(filepath.Join(ws, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write src: %v", err) }
    // Run `ami build --verbose`
    cmd := exec.Command(bin, "build", "--verbose")
    cmd.Dir = ws
    cmd.Stdin = bytes.NewReader(nil)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("ami build failed: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
    }
    // Assert LLVM debug file exists
    ll := filepath.Join(ws, "build", "debug", "llvm", "app", "u.ll")
    if _, err := os.Stat(ll); err != nil { t.Fatalf("missing llvm: %v", err) }
    // Also assert manifest.json references the LLVM path
    mf := filepath.Join(ws, "build", "debug", "manifest.json")
    b, err := os.ReadFile(mf)
    if err != nil { t.Fatalf("read manifest: %v", err) }
    var obj struct{ Packages []struct{ Name string; Units []struct{ Unit, LLVM string } } }
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    found := false
    for _, p := range obj.Packages {
        if p.Name != "app" { continue }
        for _, u := range p.Units {
            rel, _ := filepath.Rel(ws, ll)
            if u.Unit == "u" && u.LLVM == rel { found = true }
        }
    }
    if !found { t.Fatalf("manifest missing llvm path: %s", ll) }
}
