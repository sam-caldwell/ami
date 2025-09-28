package e2e

import (
    "bytes"
    "encoding/json"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

func TestE2E_AmiModList_JSON_SortedByNameThenVersion(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "mod_list", "sorted")
    cache := filepath.Join(ws, "cache")
    _ = os.RemoveAll(ws)
    // Create unsorted entries
    mustMkdir(t, filepath.Join(cache, "pkgB", "1.0.0"))
    mustMkdir(t, filepath.Join(cache, "pkgA", "2.0.0"))
    mustMkdir(t, filepath.Join(cache, "pkgA", "1.0.0"))
    mustWrite(t, filepath.Join(cache, "zzfile"), []byte("x"))

    cmd := exec.Command(bin, "mod", "list", "--json")
    cmd.Dir = ws
    absCache, _ := filepath.Abs(cache)
    cmd.Env = append(os.Environ(), "AMI_PACKAGE_CACHE="+absCache)
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil { t.Fatalf("run: %v\n%s", err, stderr.String()) }
    var res struct{ Entries []struct{ Name, Version string } }
    if err := json.Unmarshal(stdout.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, stdout.String()) }
    // expect pkgA@1.0.0, pkgA@2.0.0, pkgB@1.0.0, zzfile
    got := make([][2]string, 0, len(res.Entries))
    for _, e := range res.Entries { got = append(got, [2]string{e.Name, e.Version}) }
    want := [][2]string{{"pkgA","1.0.0"},{"pkgA","2.0.0"},{"pkgB","1.0.0"},{"zzfile",""}}
    if len(got) < len(want) { t.Fatalf("insufficient entries: %v", got) }
    for i := range want {
        if got[i][0] != want[i][0] || got[i][1] != want[i][1] {
            t.Fatalf("sorting mismatch at %d: got=%v want=%v full=%v", i, got[i], want[i], got)
        }
    }
}

func mustMkdir(t *testing.T, p string) { if err := os.MkdirAll(p, 0o755); err != nil { t.Fatalf("mkdir: %v", err) } }
func mustWrite(t *testing.T, p string, b []byte) { mustMkdir(t, filepath.Dir(p)); if err := os.WriteFile(p, b, 0o644); err != nil { t.Fatalf("write: %v", err) } }

