package root_test

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

// Reuse helper TestHelper_AmiBuild

func TestBuild_Human_MissingFile_SystemIOError(t *testing.T) {
    ws := t.TempDir()
    wsContent := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler:
    concurrency: NUM_CPU
    target: ./build
    env: []
  linker: {}
  linter: {}
packages:
  - main:
      version: 0.0.1
      root: ./src
      import: []
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil { t.Fatalf("write workspace: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    // file present but unreadable to trigger I/O error
    srcPath := filepath.Join(ws, "src", "main.ami")
    if err := os.WriteFile(srcPath, []byte("package main\n"), 0o644); err != nil { t.Fatalf("write src: %v", err) }
    if err := os.Chmod(srcPath, 0o000); err != nil { t.Fatalf("chmod: %v", err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuild")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI=1")
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err == nil { t.Fatalf("expected non-zero exit 2; output=\n%s", string(out)) }
    if ee, ok := err.(*exec.ExitError); ok {
        if code := ee.ExitCode(); code != 2 { t.Fatalf("got exit %d want 2; output=\n%s", code, string(out)) }
    } else { t.Fatalf("unexpected err type: %T", err) }
    s := strings.ToLower(string(out))
    if !strings.Contains(s, "system i/o error") && !strings.Contains(s, "failed to read") {
        t.Fatalf("expected human stderr message about I/O error; output=\n%s", string(out))
    }
}

