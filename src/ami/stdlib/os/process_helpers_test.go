package os

import (
	"errors"
	"os/exec"
	"testing"
)

func testProcError_ErrorString(t *testing.T) {
	e := &procError{"x"}
	if e.Error() != "x" {
		t.Fatalf("procError.Error mismatch: %q", e.Error())
	}
}

func testExitCodeFromError(t *testing.T) {
	if exitCodeFromError(nil) != nil {
		t.Fatalf("nil error -> nil exit code")
	}
	// Synthesize an exec.ExitError via a trivial failing command
	cmd := exec.Command("go", "tool", "unknown-subcommand-that-does-not-exist")
	err := cmd.Run()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if c := exitCodeFromError(ee); c == nil {
			t.Fatalf("expected non-nil exit code from ExitError")
		}
	}
}

func testPrestartWriter_AfterStart_Forwards(t *testing.T) {
	p, err := Exec("go", "version")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if err := p.Start(false); err != nil {
		t.Fatalf("start: %v", err)
	}
	w, err := p.Stdin()
	if err != nil {
		t.Fatalf("stdin: %v", err)
	}
	// After start, writes go to the live pipe (no-op for go version).
	if _, err := w.Write([]byte("noop")); err != nil {
		t.Fatalf("write after start: %v", err)
	}
}
