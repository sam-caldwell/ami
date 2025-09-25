package root_test

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

// Helper that runs ami --json build; enabled via GO_WANT_HELPER_AMI_JSON=1
func TestHelper_AmiBuildJSON(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_AMI_JSON") != "1" {
		return
	}
	os.Args = []string{"ami", "--json", "build"}
	code := rootcmd.Execute()
	os.Exit(code)
}

// Minimal diag record for parsing logger output
type diagRecord struct {
	Schema    string                 `json:"schema"`
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
}

func makeTaggedRepo3(t *testing.T, dir, tag string) {
	t.Helper()
	r, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	w, _ := r.Worktree()
	_, _ = w.Add("f.txt")
	h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName(tag), h)); err != nil {
		t.Fatalf("tag: %v", err)
	}
}

func TestBuild_IntegrityLogs_JSONMismatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cache := filepath.Join(home, ".ami", "pkg")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}

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
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil {
		t.Fatalf("write workspace: %v", err)
	}

	version := "v1.0.0"
	entry := filepath.Join(cache, "repo@"+version)
	if err := os.MkdirAll(entry, 0o755); err != nil {
		t.Fatalf("mkdir entry: %v", err)
	}
	makeTaggedRepo3(t, entry, version)

	// Wrong digest in ami.sum
	sum := `{"schema":"ami.sum/v1","packages":{"example/repo":{"v1.0.0":"deadbeef"}}}`
	if err := os.WriteFile(filepath.Join(ws, "ami.sum"), []byte(sum), 0o644); err != nil {
		t.Fatalf("write ami.sum: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuildJSON")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_JSON=1", "HOME="+home)
	cmd.Dir = ws
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit (3) for integrity violation; stdout=\n%s", string(out))
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if code := ee.ExitCode(); code != 3 {
			t.Fatalf("unexpected exit code: got %d want 3; stdout=\n%s", code, string(out))
		}
	} else {
		t.Fatalf("unexpected error type: %T err=%v; stdout=\n%s", err, err, string(out))
	}

	// Parse JSON lines from stdout and assert integrity error records present
	var seenMismatch bool
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var rec diagRecord
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		if rec.Schema != "diag.v1" {
			continue
		}
		if rec.Level != "error" {
			continue
		}
		switch rec.Message {
		case "integrity: digest mismatch":
			seenMismatch = true
			if rec.Data == nil {
				t.Fatalf("missing data on digest mismatch record")
			}
			if rec.Data["pkg"].(string) != "example/repo" {
				t.Fatalf("pkg mismatch: %v", rec.Data["pkg"])
			}
			if rec.Data["version"].(string) != version {
				t.Fatalf("version mismatch: %v", rec.Data["version"])
			}
		}
	}
	if !seenMismatch {
		t.Fatalf("did not observe integrity: digest mismatch record in JSON logs. stdout=\n%s", string(out))
	}
}
