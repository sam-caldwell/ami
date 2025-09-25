package root_test

import (
    "bufio"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

// captureBoth captures stdout and stderr while fn executes and returns them as strings.
func captureBoth(t *testing.T, fn func()) (string, string) {
    t.Helper()
    // stdout
    oldOut := os.Stdout
    rOut, wOut, err := os.Pipe()
    if err != nil { t.Fatalf("pipe stdout: %v", err) }
    os.Stdout = wOut
    // stderr
    oldErr := os.Stderr
    rErr, wErr, err := os.Pipe()
    if err != nil { t.Fatalf("pipe stderr: %v", err) }
    os.Stderr = wErr

    defer func() {
        os.Stdout = oldOut
        os.Stderr = oldErr
    }()

    fn()
    // Close writers to flush
    _ = wOut.Close()
    _ = wErr.Close()

    // Read back
    readAll := func(r *os.File) string {
        var b strings.Builder
        sc := bufio.NewScanner(r)
        for sc.Scan() { b.WriteString(sc.Text()); b.WriteByte('\n') }
        return b.String()
    }
    out := readAll(rOut)
    errStr := readAll(rErr)
    return out, errStr
}

func TestClean_Fresh_Human_EmitsActions(t *testing.T) {
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()
    _ = os.RemoveAll("build")

    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "clean"}

    out, errStr := captureBoth(t, func(){ _ = rootcmd.Execute() })
    if errStr != "" { t.Fatalf("unexpected stderr: %s", errStr) }
    if !strings.Contains(out, "clean.remove.skip") { t.Fatalf("missing remove.skip: %q", out) }
    if !strings.Contains(out, "clean.create") { t.Fatalf("missing create: %q", out) }
    // build dir exists
    if fi, err := os.Stat("build"); err != nil || !fi.IsDir() {
        t.Fatalf("expected build directory to exist")
    }
}

func TestClean_RemovesExistingFiles_Human(t *testing.T) {
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()
    // Setup build with a file
    if err := os.MkdirAll("build", 0o755); err != nil { t.Fatalf("mkdir build: %v", err) }
    if err := os.WriteFile(filepath.Join("build", "keep.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write file: %v", err) }

    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "clean"}

    out, _ := captureBoth(t, func(){ _ = rootcmd.Execute() })
    if !strings.Contains(out, "clean.remove") { t.Fatalf("missing remove: %q", out) }
    if !strings.Contains(out, "clean.create") { t.Fatalf("missing create: %q", out) }
    // File removed
    if _, err := os.Stat(filepath.Join("build", "keep.txt")); err == nil {
        t.Fatalf("expected file to be removed by clean")
    }
}

func TestClean_JSON_EmitsStructuredEvents(t *testing.T) {
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()
    _ = os.RemoveAll("build")

    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "--json", "clean"}

    out, errStr := captureBoth(t, func(){ _ = rootcmd.Execute() })
    if errStr != "" { t.Fatalf("unexpected stderr: %s", errStr) }

    // Expect two JSON lines: remove.skip and create
    type rec struct{ Message string; Data map[string]any }
    var msgs []string
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var r rec
        if err := json.Unmarshal([]byte(sc.Text()), &r); err != nil { t.Fatalf("bad json: %v line=%q", err, sc.Text()) }
        msgs = append(msgs, r.Message)
        if r.Data["path"] != "build" { t.Fatalf("expected path=build in data: %v", r.Data) }
    }
    have := strings.Join(msgs, ",")
    if !strings.Contains(have, "clean.remove.skip") || !strings.Contains(have, "clean.create") {
        t.Fatalf("unexpected messages: %s", have)
    }
}

func TestClean_PermissionEdge_Human_EmitsError(t *testing.T) {
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()
    if err := os.MkdirAll("build", 0o755); err != nil { t.Fatalf("mkdir build: %v", err) }
    if err := os.WriteFile(filepath.Join("build", "f.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // Remove write permission from current working directory to induce RemoveAll failure
    if err := os.Chmod(".", 0o555); err != nil { t.Fatalf("chmod cwd: %v", err) }
    defer func(){ _ = os.Chmod(".", 0o755) }()

    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "clean"}

    out, errStr := captureBoth(t, func(){ _ = rootcmd.Execute() })
    if out != "" && !strings.Contains(out, "clean.remove_failed") {
        // human errors go to stderr; stdout may be empty
    }
    if !strings.Contains(errStr, "clean.remove_failed") {
        t.Fatalf("expected error message on stderr; got: %q", errStr)
    }
    // build should still exist
    if _, err := os.Stat("build"); err != nil { t.Fatalf("expected build dir to remain on failure") }
}

func TestClean_CreatePermissionEdge_Human_EmitsError(t *testing.T) {
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()
    // Ensure no build directory exists so we hit MkdirAll
    _ = os.RemoveAll("build")
    // Remove write permission from current working directory to induce MkdirAll failure
    if err := os.Chmod(".", 0o555); err != nil { t.Fatalf("chmod cwd: %v", err) }
    defer func(){ _ = os.Chmod(".", 0o755) }()

    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "clean"}

    _, errStr := captureBoth(t, func(){ _ = rootcmd.Execute() })
    if !strings.Contains(errStr, "clean.create_failed") {
        t.Fatalf("expected create_failed message on stderr; got: %q", errStr)
    }
    if _, err := os.Stat("build"); !os.IsNotExist(err) {
        t.Fatalf("expected build directory to not exist on mkdir failure")
    }
}
