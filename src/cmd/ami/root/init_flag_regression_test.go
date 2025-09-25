package root

import (
	"bytes"
	testutil "github.com/sam-caldwell/ami/src/internal/testutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit_WithForce_DoesNotPrintHelp_AndCreatesFiles(t *testing.T) {
	ws, restore := testutil.ChdirToBuildTest(t)
	defer restore()

	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"init", "--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute returned error: %v\noutput=\n%s", err, buf.String())
	}
	out := buf.String()
	if strings.Contains(out, "Usage:") && strings.Contains(out, "help for init") {
		t.Fatalf("unexpected help output for init --force: \n%s", out)
	}
	if _, err := os.Stat(filepath.Join(ws, "ami.workspace")); err != nil {
		t.Fatalf("ami.workspace missing: %v", err)
	}
}
