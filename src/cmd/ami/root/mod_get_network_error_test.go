package root_test

import (
    "os"
    "os/exec"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

// Helper that runs ami --json mod get <url>; enabled via GO_WANT_HELPER_AMI_MOD_GET_JSON=1
func TestHelper_AmiModGetJSON(t *testing.T) {
    if os.Getenv("GO_WANT_HELPER_AMI_MOD_GET_JSON") != "1" { return }
    os.Args = []string{"ami", "--json", "mod", "get", "git+ssh://invalid.invalid/org/repo.git#v1.2.3"}
    code := rootcmd.Execute()
    os.Exit(code)
}

func TestModGet_NetworkErrorExit4(t *testing.T) {
    home := t.TempDir()
    t.Setenv("HOME", home)
    ws := t.TempDir()
    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiModGetJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_MOD_GET_JSON=1", "HOME="+home)
    cmd.Dir = ws
    _, err := cmd.CombinedOutput()
    if err == nil {
        t.Fatalf("expected exit code 4 for network error")
    }
    if ee, ok := err.(*exec.ExitError); ok {
        if code := ee.ExitCode(); code != 4 {
            t.Fatalf("unexpected exit code: got %d want 4", code)
        }
    } else {
        t.Fatalf("unexpected error type: %T err=%v", err, err)
    }
}

