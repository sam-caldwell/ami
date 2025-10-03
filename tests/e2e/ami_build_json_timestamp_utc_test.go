package e2e

import (
    "context"
    "bytes"
    "encoding/json"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "testing"
    stdtime "time"
    "github.com/sam-caldwell/ami/src/testutil"
)

func TestE2E_AmiBuild_JSON_TimestampUTC(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "build", "ts_utc")
    _ = os.RemoveAll(ws)
    mustMkdir(t, filepath.Join(ws, "src"))
    // Minimal workspace
    mustWrite(t, filepath.Join(ws, "ami.workspace"), []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n"))
    mustWrite(t, filepath.Join(ws, "src", "u.ami"), []byte("package app\nfunc F(){}\n"))
    ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(20*stdtime.Second))
    defer cancel()
    cmd := exec.CommandContext(ctx, bin, "build", "--json")
    cmd.Dir = ws
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil {
        // Build may return non-zero for diag reasons; still validate JSON record
    }
    // Expect at least one JSON line; pick the last non-empty line
    lines := bytes.Split(bytes.TrimSpace(stdout.Bytes()), []byte("\n"))
    if len(lines) == 0 { t.Fatalf("no json output: %s", stdout.String()) }
    var m map[string]any
    if err := json.Unmarshal(lines[len(lines)-1], &m); err != nil { t.Fatalf("json: %v; out=%s", err, stdout.String()) }
    ts, _ := m["timestamp"].(string)
    re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`)
    if !re.MatchString(ts) { t.Fatalf("timestamp not ISO-8601 UTC ms: %q", ts) }
}
