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

// Helper that runs ami --json mod verify; enabled via GO_WANT_HELPER_AMI_MOD_VERIFY_JSON=1
func TestHelper_AmiModVerifyJSON(t *testing.T) {
    if os.Getenv("GO_WANT_HELPER_AMI_MOD_VERIFY_JSON") != "1" { return }
    os.Args = []string{"ami", "--json", "mod", "verify"}
    code := rootcmd.Execute()
    os.Exit(code)
}

// Minimal diag record shape
type diagRecord struct {
    Schema   string                 `json:"schema"`
    Timestamp string                `json:"timestamp"`
    Level    string                 `json:"level"`
    Message  string                 `json:"message"`
    Data     map[string]interface{} `json:"data"`
}

// create a minimal git repo at dir with a tag
func makeTaggedRepo(t *testing.T, dir, tag string) {
    t.Helper()
    r, err := git.PlainInit(dir, false)
    if err != nil { t.Fatalf("init repo: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    w, _ := r.Worktree()
    _, _ = w.Add("f.txt")
    h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}})
    if err != nil { t.Fatalf("commit: %v", err) }
    if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName(tag), h)); err != nil {
        t.Fatalf("tag: %v", err)
    }
}

func TestModVerify_JSON_FailsOnMismatch(t *testing.T) {
    home := t.TempDir()
    t.Setenv("HOME", home)
    cache := filepath.Join(home, ".ami", "pkg")
    if err := os.MkdirAll(cache, 0o755); err != nil { t.Fatalf("mkdir cache: %v", err) }

    ws := t.TempDir()
    // Prepare cache entry
    version := "v1.0.0"
    base := "repo"
    entry := filepath.Join(cache, base+"@"+version)
    if err := os.MkdirAll(entry, 0o755); err != nil { t.Fatalf("mkdir entry: %v", err) }
    makeTaggedRepo(t, entry, version)

    // ami.sum with wrong digest
    sum := `{"schema":"ami.sum/v1","packages":{"example/repo":{"v1.0.0":"deadbeef"}}}`
    if err := os.WriteFile(filepath.Join(ws, "ami.sum"), []byte(sum), 0o644); err != nil { t.Fatalf("write ami.sum: %v", err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiModVerifyJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_MOD_VERIFY_JSON=1", "HOME="+home)
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err == nil {
        t.Fatalf("expected non-zero exit (3) for mod verify mismatch; stdout=\n%s", string(out))
    }
    if ee, ok := err.(*exec.ExitError); ok {
        if code := ee.ExitCode(); code != 3 {
            t.Fatalf("unexpected exit code: got %d want 3; stdout=\n%s", code, string(out))
        }
    } else {
        t.Fatalf("unexpected error type: %T err=%v; stdout=\n%s", err, err, string(out))
    }

    // Look for a digest mismatch record
    var seenMismatch bool
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var rec diagRecord
        if json.Unmarshal([]byte(sc.Text()), &rec) != nil { continue }
        if rec.Schema != "diag.v1" || rec.Level != "error" { continue }
        if rec.Message == "digest mismatch" || rec.Message == "cache entry missing" || rec.Message == "digest compute failed" {
            seenMismatch = true
            break
        }
    }
    if !seenMismatch {
        t.Fatalf("did not observe a mod verify integrity error record in JSON logs. stdout=\n%s", string(out))
    }
}
