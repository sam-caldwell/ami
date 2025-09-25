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

func TestLint_RulesFilter_OnlyCRLF(t *testing.T) {
    ws := t.TempDir()
    content := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    // multiple warnings; filter to only W_FILE_CRLF
    src := "package Main\r\nimport \"fmt\"" // CRLF + no newline + uppercase
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }
    // Use helper that accepts flags via env
    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiLintJSONWithRules")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_LINT_WITH_RULES=1", "AMI_LINT_RULES=W_FILE_CRLF", "HOME="+t.TempDir())
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("unexpected non-zero exit: %v; out=\n%s", err, string(out)) }
    // Only W_FILE_CRLF should appear
    var seenCRLF, seenOther bool
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] != "diag.v1" { continue }
        code, _ := obj["code"].(string)
        if code == "W_FILE_CRLF" { seenCRLF = true } else if code != "LINT_SUMMARY" { seenOther = true }
    }
    if !seenCRLF || seenOther { t.Fatalf("rules filter failed; out=\n%s", string(out)) }
}

func TestLint_RulesFilter_Regex(t *testing.T) {
    ws := t.TempDir()
    content := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    src := "package Main\r\nimport \"fmt\"" // W_PKG_LOWERCASE|W_FILE_*|W_UNUSED_IMPORT
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }
    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiLintJSONWithRules")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_LINT_WITH_RULES=1", "AMI_LINT_RULES=re:^W_FILE_", "HOME="+t.TempDir())
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("unexpected non-zero exit: %v; out=\n%s", err, string(out)) }
    // All diag codes (excluding summary) must start with W_FILE_
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] != "diag.v1" || obj["code"] == "LINT_SUMMARY" { continue }
        code, _ := obj["code"].(string)
        if !strings.HasPrefix(code, "W_FILE_") {
            t.Fatalf("regex filter leaked code %s; out=\n%s", code, string(out))
        }
    }
}

func TestLint_RulesFilter_Glob(t *testing.T) {
    ws := t.TempDir()
    content := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    src := "package Main\r\nimport \"fmt\"" // W_PKG_LOWERCASE|W_FILE_*|W_UNUSED_IMPORT
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }
    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiLintJSONWithRules")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_LINT_WITH_RULES=1", "AMI_LINT_RULES=W_FILE_*", "HOME="+t.TempDir())
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("unexpected non-zero exit: %v; out=\n%s", err, string(out)) }
    // All diag codes must match glob W_FILE_*
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] != "diag.v1" || obj["code"] == "LINT_SUMMARY" { continue }
        code, _ := obj["code"].(string)
        if !strings.HasPrefix(code, "W_FILE_") {
            t.Fatalf("glob filter leaked code %s; out=\n%s", code, string(out))
        }
    }
}

func TestLint_MaxWarn_LimitsEmission(t *testing.T) {
    ws := t.TempDir()
    content := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    // multiple warnings
    src := "package Main\r\nimport \"fmt\"\nimport \"fmt\"" // CRLF + no newline + uppercase + dup + unused
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }
    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiLintJSONWithRules")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_LINT_WITH_RULES=1", "AMI_LINT_MAXWARN=2", "HOME="+t.TempDir())
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("unexpected non-zero exit: %v; out=\n%s", err, string(out)) }
    // Count warn level entries (ignore summary)
    warns := 0
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] != "diag.v1" { continue }
        if obj["code"] == "LINT_SUMMARY" { continue }
        if lvl, _ := obj["level"].(string); lvl == "warn" { warns++ }
    }
    if warns > 2 { t.Fatalf("expected at most 2 warnings, got %d\n%s", warns, string(out)) }
}

func TestLint_Strict_EscalatesWarningsToErrors(t *testing.T) {
    ws := t.TempDir()
    content := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    // One warning (uppercase package)
    src := "package Main\n"
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }
    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiLintJSONWithRules")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_LINT_WITH_RULES=1", "AMI_LINT_STRICT=1", "HOME="+t.TempDir())
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err != nil { t.Fatalf("unexpected non-zero exit: %v; out=\n%s", err, string(out)) }
    // All warnings should have level=error in strict mode
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] != "diag.v1" { continue }
        if obj["code"] == "LINT_SUMMARY" { continue }
        if lvl, _ := obj["level"].(string); lvl != "error" {
            t.Fatalf("expected error level in strict mode; got: %s\n%s", lvl, string(out))
        }
    }
}
