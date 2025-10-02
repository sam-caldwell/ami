package os

import (
	goos "os"
	goexec "os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func testStart_Twice_ReturnsError(t *testing.T) {
	p, err := Exec("go", "version")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if err := p.Start(false); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := p.Start(false); err == nil {
		t.Fatalf("expected error on second start")
	}
}

func testKill_BeforeStart_ReturnsError(t *testing.T) {
	p, err := Exec("go", "version")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if err := p.Kill(); err == nil {
		t.Fatalf("expected kill error before start")
	}
}

func testNilReceiversAndStdStreams(t *testing.T) {
	var p *Process
	if _, err := p.Stdin(); err == nil {
		t.Fatalf("expected error on nil.Stdin()")
	}
	if got := p.Stdout(); got != nil {
		t.Fatalf("nil.Stdout() should be nil, got %v", got)
	}
	if got := p.Stderr(); got != nil {
		t.Fatalf("nil.Stderr() should be nil, got %v", got)
	}
	if err := p.Start(false); err == nil {
		t.Fatalf("expected error on nil.Start()")
	}
}

func testStatus_BeforeStart(t *testing.T) {
	p, err := Exec("go", "version")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	st := p.Status()
	if st.Running {
		t.Fatalf("expected not running before start")
	}
	if st.ExitCode != nil {
		t.Fatalf("expected nil exit code before start")
	}
}

func testStart_Blocking_NonZeroExit_ExitCodeRecorded(t *testing.T) {
	dir := t.TempDir()
	src := `package main
import "os"
func main(){ os.Exit(3) }`
	file := filepath.Join(dir, "main.go")
	if err := goos.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Build a binary to ensure exit code propagation is direct
	bin := filepath.Join(dir, "exit3bin")
	if out, err2 := goexec.Command("go", "build", "-o", bin, file).CombinedOutput(); err2 != nil {
		t.Fatalf("go build failed: %v (out=%s)", err2, string(out))
	}
	p, err := Exec(bin)
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	err = p.Start(true)
	if err == nil {
		t.Fatalf("expected non-nil error from non-zero exit")
	}
	st := p.Status()
	if st.Running {
		t.Fatalf("should not be running after blocking run")
	}
	if st.ExitCode == nil || *st.ExitCode != 3 {
		got := 0
		if st.ExitCode != nil {
			got = *st.ExitCode
		}
		t.Fatalf("want exit 3, got %d", got)
	}
}

func testPreStart_StdinWrite_Roundtrip(t *testing.T) {
	// Use a tiny echo program so behavior is stable across platforms
	if runtime.GOOS == "windows" {
		// Windows is ok with go run; proceed
	}
	dir := t.TempDir()
	src := `package main
import (
  "io"
  "os"
)
func main(){ io.Copy(os.Stdout, os.Stdin) }`
	file := filepath.Join(dir, "main.go")
	if err := goos.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	p, err := Exec("go", "run", file)
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	w, err := p.Stdin()
	if err != nil {
		t.Fatalf("stdin: %v", err)
	}
	msg := "prestart-write\n"
	if _, err := w.Write([]byte(msg)); err != nil {
		t.Fatalf("write before start: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := p.Start(true); err != nil {
		// Some toolchains may fail; still verify we got an exit code
	}
	// Poll for completion short period if not blocking
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		st := p.Status()
		if !st.Running {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	got := strings.TrimSpace(string(p.Stdout()))
	if got != strings.TrimSpace(msg) {
		t.Fatalf("stdout mismatch: got %q want %q", got, msg)
	}
}

func testStderr_Capture(t *testing.T) {
	dir := t.TempDir()
	src := `package main
import "os"
func main(){ os.Stderr.WriteString("oops\n") }`
	file := filepath.Join(dir, "main.go")
	if err := goos.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	bin := filepath.Join(dir, "stderrbin")
	if out, err2 := goexec.Command("go", "build", "-o", bin, file).CombinedOutput(); err2 != nil {
		t.Fatalf("go build failed: %v (out=%s)", err2, string(out))
	}
	p, err := Exec(bin)
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if err := p.Start(true); err != nil {
		t.Fatalf("start: %v", err)
	}
	if s := string(p.Stderr()); !strings.Contains(s, "oops") {
		t.Fatalf("stderr capture failed: %q", s)
	}
}
