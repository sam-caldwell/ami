package root_test

import (
    "bufio"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    ammod "github.com/sam-caldwell/ami/src/ami/mod"
    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

// captureBoth captures stdout and stderr while fn executes and returns them.
func captureBothMod(t *testing.T, fn func()) (string, string) {
    t.Helper()
    oldOut := os.Stdout
    rOut, wOut, _ := os.Pipe()
    os.Stdout = wOut
    oldErr := os.Stderr
    rErr, wErr, _ := os.Pipe()
    os.Stderr = wErr
    defer func(){ os.Stdout = oldOut; os.Stderr = oldErr }()
    fn()
    _ = wOut.Close(); _ = wErr.Close()
    readAll := func(r *os.File) string { var b strings.Builder; sc:=bufio.NewScanner(r); for sc.Scan(){b.WriteString(sc.Text());b.WriteByte('\n')} ; return b.String() }
    return readAll(rOut), readAll(rErr)
}

func TestModClean_Fresh_Human(t *testing.T) {
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()
    // Isolated HOME
    home := t.TempDir()
    t.Setenv("HOME", home)
    dir, err := ammod.CacheDirPath()
    if err != nil { t.Fatalf("cache path: %v", err) }
    _ = os.RemoveAll(filepath.Dir(dir))

    old := os.Args; defer func(){ os.Args = old }()
    os.Args = []string{"ami", "mod", "clean"}
    out, errStr := captureBothMod(t, func(){ _ = rootcmd.Execute() })
    if errStr != "" { t.Fatalf("unexpected stderr: %s", errStr) }
    if !strings.Contains(out, "cache.remove.skip") || !strings.Contains(out, "cache.create") {
        t.Fatalf("unexpected output: %q", out)
    }
    if fi, err := os.Stat(dir); err != nil || !fi.IsDir() { t.Fatalf("cache dir not created: %s", dir) }
}

func TestModClean_Existing_Human(t *testing.T) {
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()
    home := t.TempDir(); t.Setenv("HOME", home)
    dir, err := ammod.CacheDir() // creates dir
    if err != nil { t.Fatalf("cache dir: %v", err) }
    // Put a file inside
    if err := os.WriteFile(filepath.Join(dir, "x.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    old := os.Args; defer func(){ os.Args = old }()
    os.Args = []string{"ami", "mod", "clean"}
    out, _ := captureBothMod(t, func(){ _ = rootcmd.Execute() })
    if !strings.Contains(out, "cache.remove") || !strings.Contains(out, "cache.create") {
        t.Fatalf("unexpected output: %q", out)
    }
    if _, err := os.Stat(filepath.Join(dir, "x.txt")); err == nil { t.Fatalf("expected file removed by mod clean") }
}

func TestModClean_JSON(t *testing.T) {
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()
    home := t.TempDir(); t.Setenv("HOME", home)
    dir, err := ammod.CacheDirPath(); if err != nil { t.Fatalf("path: %v", err) }
    _ = os.RemoveAll(filepath.Dir(dir))

    old := os.Args; defer func(){ os.Args = old }()
    os.Args = []string{"ami", "--json", "mod", "clean"}
    out, errStr := captureBothMod(t, func(){ _ = rootcmd.Execute() })
    if errStr != "" { t.Fatalf("unexpected stderr: %s", errStr) }
    type rec struct{ Message string; Data map[string]any }
    msgs := []string{}
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var r rec
        if err := json.Unmarshal([]byte(sc.Text()), &r); err != nil { t.Fatalf("invalid json: %v line=%q", err, sc.Text()) }
        if r.Data["path"] == "" { t.Fatalf("expected path in data: %v", r.Data) }
        msgs = append(msgs, r.Message)
    }
    joined := strings.Join(msgs, ",")
    if !strings.Contains(joined, "cache.remove.skip") || !strings.Contains(joined, "cache.create") {
        t.Fatalf("unexpected messages: %s", joined)
    }
}

func TestModClean_PermissionEdge_Error(t *testing.T) {
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()
    home := t.TempDir(); t.Setenv("HOME", home)
    dir, err := ammod.CacheDir() // creates ~/.ami/pkg
    if err != nil { t.Fatalf("cache dir: %v", err) }
    // Restrict parent (~/.ami) permissions to prevent removal of pkg
    parent := filepath.Dir(dir)
    if err := os.Chmod(parent, 0o555); err != nil { t.Fatalf("chmod parent: %v", err) }
    defer func(){ _ = os.Chmod(parent, 0o755) }()

    old := os.Args; defer func(){ os.Args = old }()
    os.Args = []string{"ami", "mod", "clean"}
    _, errStr := captureBothMod(t, func(){ _ = rootcmd.Execute() })
    if !strings.Contains(errStr, "cache.remove_failed") {
        t.Fatalf("expected cache.remove_failed on stderr; got: %q", errStr)
    }
    // ensure directory still exists
    if _, e := os.Stat(dir); e != nil { t.Fatalf("expected cache dir to remain on failure") }
}
