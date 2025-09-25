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

// Sad path: mismatch between constraint and local package version
func TestLint_CrossPackageConstraint_Mismatch(t *testing.T) {
    ws := t.TempDir()
    wsContent := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: ["lib/util >=1.0.0"] }
  - lib/util: { version: 0.1.0, root: ./lib/util, import: [] }
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "lib", "util"), 0o755); err != nil { t.Fatalf("mkdir lib/util: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "lib", "util", "util.ami"), []byte("package util\n"), 0o644); err != nil { t.Fatalf("write util src: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte("package main\nimport \"lib/util\"\n"), 0o644); err != nil { t.Fatalf("write main: %v", err) }
    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiLintJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_JSON=1", "HOME="+t.TempDir())
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("unexpected non-zero exit: %v; out=\n%s", err, string(out)) }
    var seen bool
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "E_IMPORT_CONSTRAINT" { seen = true; break }
    }
    if !seen { t.Fatalf("expected E_IMPORT_CONSTRAINT; out=\n%s", string(out)) }
}

// Happy path: matching constraint
func TestLint_CrossPackageConstraint_Match(t *testing.T) {
    ws := t.TempDir()
    wsContent := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: ["lib/util >=0.1.0"] }
  - lib/util: { version: 0.1.0, root: ./lib/util, import: [] }
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "lib", "util"), 0o755); err != nil { t.Fatalf("mkdir lib/util: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "lib", "util", "util.ami"), []byte("package util\n"), 0o644); err != nil { t.Fatalf("write util src: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte("package main\nimport \"lib/util\"\n"), 0o644); err != nil { t.Fatalf("write main: %v", err) }
    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiLintJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_JSON=1", "HOME="+t.TempDir())
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("unexpected non-zero exit: %v; out=\n%s", err, string(out)) }
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "E_IMPORT_CONSTRAINT" {
            t.Fatalf("did not expect E_IMPORT_CONSTRAINT; out=\n%s", string(out))
        }
    }
}

