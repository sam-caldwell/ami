package e2e

import (
    "bytes"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

func TestE2E_AmiBuild_Verbose_LogsToolchainVersionOrMissing(t *testing.T) {
    bin := buildAmi(t)
    dir := filepath.Join("build", "test", "e2e", "build", "tool_ver")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n"), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "main.ami"), []byte("package app\n"), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    cmd := exec.Command(bin, "build", "--verbose")
    cmd.Dir = dir
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    _ = cmd.Run()

    data, err := os.ReadFile(filepath.Join(dir, "build", "debug", "activity.log"))
    if err != nil { t.Fatalf("read activity.log: %v", err) }
    s := string(data)
    if !bytes.Contains(data, []byte("\"message\":\"toolchain.clang\"")) &&
        !bytes.Contains(data, []byte("\"message\":\"toolchain.missing\"")) {
        t.Fatalf("expected toolchain version or missing log; got: %s", s)
    }
}
