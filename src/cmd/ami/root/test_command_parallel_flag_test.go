package root_test

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

func TestTest_JSON_ParallelFlagAccepted(t *testing.T) {
    ws := t.TempDir()
    gomod := "module example.com/ami-test\n\ngo 1.22\n"
    if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte(gomod), 0o644); err != nil { t.Fatalf("write go.mod: %v", err) }
    src := `package main
import "testing"
func TestA(t *testing.T){ }
func TestB(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(ws, "a_test.go"), []byte(src), 0o644); err != nil { t.Fatalf("write test: %v", err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiTestJSON_Parallel1")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_TEST_JSON_PARALLEL=1")
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("unexpected non-zero exit: %v, stdout=\n%s", err, string(out)) }
}

