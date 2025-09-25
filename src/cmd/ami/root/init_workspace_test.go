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

func TestInit_CreatesNewWorkspaceScaffold(t *testing.T) {
    ws, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    old := os.Args
    defer func(){ os.Args = old }()
    os.Args = []string{"ami", "init", "--force"}
    _ = rootcmd.Execute()

    // Files exist
    if _, err := os.Stat(filepath.Join(ws, "ami.workspace")); err != nil {
        t.Fatalf("ami.workspace missing: %v", err)
    }
    if _, err := os.Stat(filepath.Join(ws, "src", "main.ami")); err != nil {
        t.Fatalf("src/main.ami missing: %v", err)
    }
    if _, err := os.Stat(filepath.Join(ws, ".gitignore")); err != nil {
        t.Fatalf(".gitignore missing: %v", err)
    }

    // Scaffold content sanity
    b, _ := os.ReadFile(filepath.Join(ws, "ami.workspace"))
    s := string(b)
    if !strings.Contains(s, "version: 1.0.0") || !strings.Contains(s, "toolchain:") || !strings.Contains(s, "packages:") {
        t.Fatalf("unexpected ami.workspace content: %q", s)
    }
    m, _ := os.ReadFile(filepath.Join(ws, "src", "main.ami"))
    if !strings.Contains(string(m), "AMI main entrypoint (scaffold)") {
        t.Fatalf("unexpected main.ami content: %q", string(m))
    }
    gi, _ := os.ReadFile(filepath.Join(ws, ".gitignore"))
    if !strings.Contains(string(gi), "./build") {
        t.Fatalf(".gitignore missing ./build entry: %q", string(gi))
    }
}

func TestInit_Reinit_Idempotence_WithForce(t *testing.T) {
    ws, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    run := func() {
        old := os.Args
        defer func(){ os.Args = old }()
        os.Args = []string{"ami", "init", "--force"}
        _ = rootcmd.Execute()
    }

    run()
    w1, _ := os.ReadFile(filepath.Join(ws, "ami.workspace"))
    m1, _ := os.ReadFile(filepath.Join(ws, "src", "main.ami"))
    _, _ = os.ReadFile(filepath.Join(ws, ".gitignore"))

    run()
    w2, _ := os.ReadFile(filepath.Join(ws, "ami.workspace"))
    m2, _ := os.ReadFile(filepath.Join(ws, "src", "main.ami"))
    gi2, _ := os.ReadFile(filepath.Join(ws, ".gitignore"))

    if string(w1) != string(w2) {
        t.Fatalf("ami.workspace not idempotent across reinit with --force")
    }
    if string(m1) != string(m2) {
        t.Fatalf("src/main.ami not idempotent across reinit with --force")
    }
    if strings.Count(string(gi2), "./build") != 1 {
        t.Fatalf(".gitignore contains duplicate ./build entries: %q", string(gi2))
    }
}

func TestInit_JSON_Output(t *testing.T) {
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    old := os.Args
    defer func(){ os.Args = old }()
    os.Args = []string{"ami", "--json", "init", "--force"}

    out := captureStdout(t, func(){ _ = rootcmd.Execute() })
    type rec struct {
        Schema  string                 `json:"schema"`
        Message string                 `json:"message"`
        Data    map[string]interface{} `json:"data"`
    }
    var msgs []string
    var workSeen, sourceSeen bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var r rec
        if err := json.Unmarshal([]byte(sc.Text()), &r); err != nil {
            t.Fatalf("invalid JSON line: %v line=%q", err, sc.Text())
        }
        if r.Schema == "diag.v1" {
            msgs = append(msgs, r.Message)
        }
        if r.Message == "workspace initialized" {
            if v, ok := r.Data["workspace"].(string); ok && v == "ami.workspace" { workSeen = true }
            if v, ok := r.Data["source"].(string); ok && v == filepath.Join("src","main.ami") { sourceSeen = true }
        }
    }
    got := strings.Join(msgs, ",")
    if !strings.Contains(got, "initialized git repository") {
        t.Fatalf("expected init git message in JSON: %s", got)
    }
    if !strings.Contains(got, "workspace initialized") {
        t.Fatalf("expected workspace initialized message in JSON: %s", got)
    }
    if !workSeen || !sourceSeen {
        t.Fatalf("expected workspace/source fields in JSON for workspace initialized")
    }
}

func TestInit_JSON_Error_WhenNotGitRepoWithoutForce(t *testing.T) {
    ws, restore := testutil.ChdirToBuildTest(t)
    defer restore()
    // Sanity: ensure fresh workspace without .git
    if _, err := os.Stat(filepath.Join(ws, ".git")); !os.IsNotExist(err) {
        t.Fatalf("precondition: expected no .git directory in temp workspace")
    }

    old := os.Args
    defer func(){ os.Args = old }()
    os.Args = []string{"ami", "--json", "init"}

    out, errStr := captureBoth(t, func(){ _ = rootcmd.Execute() })
    if errStr != "" {
        t.Fatalf("unexpected stderr in JSON mode: %q", errStr)
    }

    // Expect a JSON error record on stdout and no files created
    type rec struct {
        Schema  string `json:"schema"`
        Level   string `json:"level"`
        Message string `json:"message"`
    }
    var found bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var r rec
        if err := json.Unmarshal([]byte(sc.Text()), &r); err != nil {
            t.Fatalf("invalid JSON line: %v line=%q", err, sc.Text())
        }
        if r.Schema == "diag.v1" && r.Level == "error" && strings.Contains(r.Message, "not a git repository") {
            found = true
        }
    }
    if !found {
        t.Fatalf("expected JSON error record indicating not a git repository; got: %q", out)
    }
    if _, err := os.Stat(filepath.Join(ws, "ami.workspace")); err == nil {
        t.Fatalf("did not expect ami.workspace to be created on error")
    }
    if _, err := os.Stat(filepath.Join(ws, "src", "main.ami")); err == nil {
        t.Fatalf("did not expect src/main.ami to be created on error")
    }
}
