package e2e

import (
    "bytes"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "testing"
)

// When clang is present, ami build should place the binary under build/<env>/.
func TestE2E_AmiBuild_EnvBinaryPath(t *testing.T) {
    if _, err := exec.LookPath("clang"); err != nil { t.Skip("clang missing") }
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "build", "env_bin")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n"), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write src: %v", err) }
    cmd := exec.Command(bin, "build")
    cmd.Dir = ws
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    _ = cmd.Run()
    env := runtime.GOOS + "/" + runtime.GOARCH
    // Look for binary under build/<env>/app
    target := filepath.Join(ws, "build", env, "app")
    if st, err := os.Stat(target); err != nil || st.IsDir() { t.Fatalf("expected binary at %s; err=%v st=%v", target, err, st) }
}

