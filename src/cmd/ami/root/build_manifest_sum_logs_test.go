package root_test

import (
    "bufio"
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

// Focused test: existing ami.manifest vs ami.sum mismatch causes exit 3 and logs diag
func TestBuild_ExistingManifestVsSum_JSONMismatch(t *testing.T) {
    home := t.TempDir()
    t.Setenv("HOME", home)

    ws := t.TempDir()
    wsContent := `version: 1.0.0
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
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil { t.Fatalf("write workspace: %v", err) }

    // ami.sum expects digest "deadbeef"
    sum := `{"schema":"ami.sum/v1","packages":{"example/repo":{"v1.0.0":"deadbeef"}}}`
    if err := os.WriteFile(filepath.Join(ws, "ami.sum"), []byte(sum), 0o644); err != nil { t.Fatalf("write ami.sum: %v", err) }

    // existing ami.manifest has mismatched digest "beefdead"
    man := `{
  "schema": "ami.manifest/v1",
  "project": {"name": "demo", "version": "0.0.1"},
  "packages": [
    {"name": "example/repo", "version": "v1.0.0", "digestSHA256": "beefdead", "source": "/x"}
  ],
  "artifacts": [],
  "toolchain": {"amiVersion": "v0.0.0-dev", "goVersion": "1.25"},
  "createdAt": "2025-01-01T00:00:00.000Z"
}`
    if err := os.WriteFile(filepath.Join(ws, "ami.manifest"), []byte(man), 0o644); err != nil { t.Fatalf("write ami.manifest: %v", err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuildJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_JSON=1", "HOME="+home)
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err == nil {
        t.Fatalf("expected non-zero exit (3) for manifest vs sum mismatch; stdout=\n%s", string(out))
    }
    if ee, ok := err.(*exec.ExitError); ok {
        if code := ee.ExitCode(); code != 3 {
            t.Fatalf("unexpected exit code: got %d want 3; stdout=\n%s", code, string(out))
        }
    } else {
        t.Fatalf("unexpected error type: %T err=%v; stdout=\n%s", err, err, string(out))
    }

    // diag line shape for parsing
    type diagLine struct {
        Schema   string                 `json:"schema"`
        Timestamp string                `json:"timestamp"`
        Level    string                 `json:"level"`
        Message  string                 `json:"message"`
        Data     map[string]interface{} `json:"data"`
    }

    // Look for the specific error message in JSON logs
    var seen bool
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var rec diagLine
        line := strings.TrimSpace(sc.Text())
        if line == "" { continue }
        if json.Unmarshal([]byte(line), &rec) != nil { continue }
        if rec.Schema != "diag.v1" { continue }
        if rec.Level != "error" { continue }
        if rec.Message == "integrity: existing manifest vs ami.sum mismatch" {
            seen = true
            break
        }
    }
    if !seen {
        t.Fatalf("did not observe 'integrity: existing manifest vs ami.sum mismatch' record in JSON logs. stdout=\n%s", string(out))
    }
}

