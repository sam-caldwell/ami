package root_test

import (
    "bufio"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

type diagRecord struct {
    Schema    string                 `json:"schema"`
    Timestamp string                 `json:"timestamp"`
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Data      map[string]interface{} `json:"data"`
}

// captureStdout captures stdout while fn executes and returns captured output as string.
func captureStdout(t *testing.T, fn func()) string {
    t.Helper()
    old := os.Stdout
    r, w, err := os.Pipe()
    if err != nil { t.Fatalf("pipe: %v", err) }
    os.Stdout = w
    defer func() { os.Stdout = old }()
    fn()
    w.Close()
    var b strings.Builder
    sc := bufio.NewScanner(r)
    for sc.Scan() { b.WriteString(sc.Text()); b.WriteByte('\n') }
    return b.String()
}

func TestModList_JSON_IncludesDigestFromAmiSum(t *testing.T) {
    // Setup isolated HOME and workspace
    tmp := t.TempDir()
    // HOME for cache
    t.Setenv("HOME", tmp)
    cacheDir := filepath.Join(tmp, ".ami", "pkg")
    if err := os.MkdirAll(cacheDir, 0o755); err != nil { t.Fatalf("mkdir cache: %v", err) }
    // Cached entries
    if err := os.MkdirAll(filepath.Join(cacheDir, "repo@v1.2.3"), 0o755); err != nil { t.Fatalf("mkdir entry: %v", err) }
    if err := os.MkdirAll(filepath.Join(cacheDir, "other@v0.1.0"), 0o755); err != nil { t.Fatalf("mkdir entry: %v", err) }

    // Workspace with ami.sum
    ws := t.TempDir()
    if err := os.WriteFile(filepath.Join(ws, "ami.sum"), []byte(`{
  "schema": "ami.sum/v1",
  "packages": {
    "github.com/example/repo": {"v1.2.3": "deadbeefcafebabe"}
  }
}`), 0o644); err != nil { t.Fatalf("write ami.sum: %v", err) }
    // Run from workspace directory
    cwd, _ := os.Getwd()
    defer os.Chdir(cwd)
    _ = os.Chdir(ws)

    // Execute CLI: ami --json mod list
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "--json", "mod", "list"}

    out := captureStdout(t, func() {
        _ = rootcmd.Execute()
    })

    // Parse JSON lines and verify digest present for matching entry
    var seenRepo, seenOther bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var rec diagRecord
        if err := json.Unmarshal([]byte(sc.Text()), &rec); err != nil {
            t.Fatalf("invalid json line: %v\nline: %s", err, sc.Text())
        }
        if rec.Message != "cache.entry" {
            // Older human output should not appear in --json mode
            t.Fatalf("unexpected message: %s", rec.Message)
        }
        if rec.Data == nil { t.Fatalf("missing data object") }
        entry, _ := rec.Data["entry"].(string)
        switch entry {
        case "repo@v1.2.3":
            seenRepo = true
            if _, ok := rec.Data["digest"]; !ok {
                t.Fatalf("expected digest for %s", entry)
            }
            if rec.Data["digest"].(string) != "deadbeefcafebabe" {
                t.Fatalf("unexpected digest: %v", rec.Data["digest"])
            }
        case "other@v0.1.0":
            seenOther = true
            if _, ok := rec.Data["digest"]; ok {
                t.Fatalf("did not expect digest for %s", entry)
            }
        }
    }
    if !seenRepo || !seenOther {
        t.Fatalf("expected both entries listed; seenRepo=%v seenOther=%v", seenRepo, seenOther)
    }
}

