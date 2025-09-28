package e2e

import (
    "bytes"
    "encoding/json"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "testing"
)

func TestE2E_AmiModList_JSON_TimestampsUTC(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "mod_list", "ts_utc")
    cache := filepath.Join(ws, "cache")
    _ = os.RemoveAll(ws)
    // Create entries
    mustMkdir(t, filepath.Join(cache, "pkgA", "1.2.3"))
    mustWrite(t, filepath.Join(cache, "pkgA", "1.2.3", "f.txt"), []byte("x"))
    mustWrite(t, filepath.Join(cache, "fileB"), []byte("y"))

    cmd := exec.Command(bin, "mod", "list", "--json")
    cmd.Dir = ws
    absCache, _ := filepath.Abs(cache)
    cmd.Env = append(os.Environ(), "AMI_PACKAGE_CACHE="+absCache)
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil { t.Fatalf("run: %v\n%s", err, stderr.String()) }
    var res struct{ Entries []struct{ Modified string } }
    if err := json.Unmarshal(stdout.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, stdout.String()) }
    re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z$`)
    for _, e := range res.Entries {
        if !re.MatchString(e.Modified) {
            t.Fatalf("modified not ISO-8601 UTC: %q", e.Modified)
        }
    }
}

