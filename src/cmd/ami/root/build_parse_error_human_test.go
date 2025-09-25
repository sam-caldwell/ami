package root_test

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

// Reuse TestHelper_AmiBuild from build_integrity_fail_test.go

func TestBuild_Human_ParseError_StderrAndExit1(t *testing.T) {
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
    // invalid package identifier
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte("package 123\n"), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuild")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI=1")
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err == nil { t.Fatalf("expected non-zero exit for parse error; stdout/stderr=\n%s", string(out)) }
    if ee, ok := err.(*exec.ExitError); ok {
        if code := ee.ExitCode(); code != 1 {
            t.Fatalf("unexpected exit code: got %d want 1; output=\n%s", code, string(out))
        }
    } else { t.Fatalf("unexpected err type: %T", err) }
    if !strings.Contains(strings.ToLower(string(out)), "invalid package identifier") {
        t.Fatalf("expected human stderr to contain parser error; output=\n%s", string(out))
    }
}

