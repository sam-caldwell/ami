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

// Reuses TestHelper_AmiBuildJSON from build_integrity_logs_test.go

func TestBuild_WorkspaceJSONDiagnostics_OnSchemaViolation(t *testing.T) {
    home := t.TempDir()
    t.Setenv("HOME", home)

    ws := t.TempDir()
    // invalid target: absolute path should be rejected
    wsContent := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler:
    concurrency: 1
    target: /abs
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

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuildJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_JSON=1", "HOME="+home)
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err == nil {
        t.Fatalf("expected exit code 1 for user error; stdout=\n%s", string(out))
    }
    if ee, ok := err.(*exec.ExitError); ok {
        if code := ee.ExitCode(); code != 1 {
            t.Fatalf("unexpected exit code: got %d want 1; stdout=\n%s", code, string(out))
        }
    } else {
        t.Fatalf("unexpected error type: %T err=%v; stdout=\n%s", err, err, string(out))
    }

    // Expect a diag.v1 JSON line with our schema error
    type diag struct {
        Schema string `json:"schema"`
        Level  string `json:"level"`
        Code   string `json:"code"`
        Message string `json:"message"`
        File   string `json:"file"`
    }
    var seen bool
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        line := strings.TrimSpace(sc.Text())
        if line == "" { continue }
        var d diag
        if json.Unmarshal([]byte(line), &d) != nil { continue }
        if d.Schema == "diag.v1" && d.Level == "error" && d.Code == "E_WS_SCHEMA" && strings.Contains(d.Message, "workspace validation failed") {
            if !strings.HasSuffix(d.File, "ami.workspace") { t.Fatalf("diag file not set to ami.workspace: %q", d.File) }
            seen = true
            break
        }
    }
    if !seen {
        t.Fatalf("did not observe diag.v1 JSON for workspace schema error. stdout=\n%s", string(out))
    }
}

