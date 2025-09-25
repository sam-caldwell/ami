package root_test

import (
    "bufio"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    sch "github.com/sam-caldwell/ami/src/schemas"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

func TestBuild_EmitsBuildPlan_HumanVerbose(t *testing.T) {
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    wsContent := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
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
    if err := os.WriteFile("ami.workspace", []byte(wsContent), 0o644); err != nil { t.Fatalf("write workspace: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte("package main\n"), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    // capture stdout and run build --verbose
    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "build", "--verbose"}
    out := captureStdoutBuild(t, func(){ _ = rootcmd.Execute() })

    // plan file exists and validates
    planPath := filepath.Join("build","debug","buildplan.json")
    b, err := os.ReadFile(planPath)
    if err != nil { t.Fatalf("missing buildplan.json: %v", err) }
    var plan sch.BuildPlanV1
    if err := json.Unmarshal(b, &plan); err != nil { t.Fatalf("unmarshal plan: %v", err) }
    if err := plan.Validate(); err != nil { t.Fatalf("plan validate: %v", err) }
    if len(plan.Targets) == 0 { t.Fatalf("expected at least one target in plan") }
    // human log mentions plan write
    if !strings.Contains(out, "build plan written:") {
        t.Fatalf("expected human log to mention plan write; got:\n%s", out)
    }
}

func TestBuild_EmitsBuildPlan_JSON(t *testing.T) {
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    wsContent := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
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
    if err := os.WriteFile("ami.workspace", []byte(wsContent), 0o644); err != nil { t.Fatalf("write workspace: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte("package main\n"), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    // Run ami --json build via helper
    old := os.Getenv("GO_WANT_HELPER_AMI_JSON")
    t.Setenv("GO_WANT_HELPER_AMI_JSON", "1")
    defer t.Setenv("GO_WANT_HELPER_AMI_JSON", old)

    // Simulate the helper flow inline without spawning a new process
    // Capture stdout
    oldArgs := os.Args
    out := captureStdoutBuild(t, func(){
        os.Args = []string{"ami", "--json", "build"}
        _ = rootcmd.Execute()
    })
    os.Args = oldArgs
    // Scan for a buildplan.v1 line
    var seen bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        line := strings.TrimSpace(sc.Text())
        if line == "" { continue }
        var probe map[string]any
        if json.Unmarshal([]byte(line), &probe) != nil { continue }
        if probe["schema"] == "buildplan.v1" {
            var plan sch.BuildPlanV1
            if json.Unmarshal([]byte(line), &plan) == nil {
                if plan.Validate() == nil && len(plan.Targets) >= 1 { seen = true; break }
            }
        }
    }
    if !seen { t.Fatalf("did not observe buildplan.v1 JSON in output. got:\n%s", out) }
}
