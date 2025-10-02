package os

import (
	goos "os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func testProcess_StdinStdout_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	src := `package main
import (
  "io"
  "os"
)
func main(){
  io.Copy(os.Stdout, os.Stdin)
}`
	file := filepath.Join(dir, "main.go")
	if err := goos.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	p, err := Exec("go", "run", file)
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
	msg := "hello world\n"
	if _, err := w.Write([]byte(msg)); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	_ = w.Close()
	// wait for completion
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		st := p.Status()
		if !st.Running && st.ExitCode != nil {
			got := string(p.Stdout())
			if strings.TrimSpace(got) != strings.TrimSpace(msg) {
				t.Fatalf("stdout mismatch: got %q want %q", got, msg)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for process to exit; status=%+v", p.Status())
}

func testProcess_Kill_StopsProcess(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("kill semantics check is validated on linux; skipping on this platform")
	}
	dir := t.TempDir()
	src := `package main
import (
  "time"
)
func main(){
  time.Sleep(10 * time.Second)
}`
	file := filepath.Join(dir, "main.go")
	if err := goos.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	bin := filepath.Join(dir, "sleepbin")
	// Build the binary to avoid killing only the go tool process
	if out, err := exec.Command("go", "build", "-o", bin, file).CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v (out=%s)", err, string(out))
	}
	p, err := Exec(bin)
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if err := p.Start(false); err != nil {
		t.Fatalf("start: %v", err)
	}
	if p.Pid() <= 0 {
		t.Fatalf("expected pid > 0 after start")
	}
	// Give the process a moment to start
	time.Sleep(50 * time.Millisecond)
	if err := p.Kill(); err != nil {
		t.Fatalf("kill: %v", err)
	}
	// wait for status to reflect exit
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		st := p.Status()
		if !st.Running {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("process still running after kill; status=%+v", p.Status())
}
