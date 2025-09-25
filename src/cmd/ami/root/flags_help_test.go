package root_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

// Helper that runs ami --json --color version to exercise mutual exclusion; enabled via GO_WANT_HELPER_AMI_JSONCOLOR=1
func TestHelper_AmiJSONColor(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_AMI_JSONCOLOR") != "1" {
		return
	}
	os.Args = []string{"ami", "--json", "--color", "version"}
	code := rootcmd.Execute()
	os.Exit(code)
}

func TestRoot_JSONAndColor_MutuallyExclusive_Exit1(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiJSONColor")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_JSONCOLOR=1")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected exit 1 for --json --color; output=\n%s", string(out))
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if code := ee.ExitCode(); code != 1 {
			t.Fatalf("unexpected exit code: got %d want 1; output=\n%s", code, string(out))
		}
	} else {
		t.Fatalf("unexpected error type: %T err=%v", err, err)
	}
	// Expect plain text error (not JSON) mentioning the conflict
	if !strings.Contains(string(out), "--json and --color cannot be used together") {
		t.Fatalf("expected conflict message; got=\n%s", string(out))
	}
}

func TestRoot_Help_PrintsUsage(t *testing.T) {
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"ami", "--help"}
	out := captureStdout(t, func() { _ = rootcmd.Execute() })
	if !strings.Contains(out, "Usage:") && !strings.Contains(out, "AMI CLI") {
		t.Fatalf("expected help/usage text; got=\n%s", out)
	}
}
