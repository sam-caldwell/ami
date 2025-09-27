package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "regexp"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Human summary line should reflect same counts as JSON summary.
func TestLint_Human_JSON_SummaryParity(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "human_json_parity")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\n// TODO: a\n// TODO: b\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{} // non-strict
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    var js bytes.Buffer
    if err := runLint(&js, dir, true, false, false); err != nil { /* warnings allowed */ }
    // Parse JSON summary
    dec := json.NewDecoder(&js)
    var errors, warnings int
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "SUMMARY" {
            data := m["data"].(map[string]any)
            errors = int(data["errors"].(float64))
            warnings = int(data["warnings"].(float64))
            break
        }
    }
    if warnings < 2 { t.Fatalf("expected >=2 warnings, got %d", warnings) }
    // Human
    var hm bytes.Buffer
    if err := runLint(&hm, dir, false, false, false); err != nil { t.Fatalf("human runLint: %v", err) }
    re := regexp.MustCompile(`lint: \d+ error\(s\), \d+ warning\(s\)\n`)
    if !re.Match(hm.Bytes()) { t.Fatalf("missing human summary line: %s", hm.String()) }
    // Basic parity: ensure exact numbers appear
    want := []byte(hmSummary(errors, warnings))
    if !bytes.Contains(hm.Bytes(), want) {
        t.Fatalf("human summary does not reflect JSON counts; want contains: %q; got: %s", string(want), hm.String())
    }
}

func hmSummary(errors, warnings int) string {
    return "lint: " + itos(errors) + " error(s), " + itos(warnings) + " warning(s)\n"
}

func itos(x int) string {
    if x == 0 { return "0" }
    b := []byte{}
    n := x
    if n < 0 { b = append(b, '-'); n = -n }
    digits := []byte{}
    for n > 0 { digits = append(digits, byte('0'+n%10)); n /= 10 }
    for i := len(digits)-1; i >= 0; i-- { b = append(b, digits[i]) }
    return string(b)
}

// Human mode: max-warn should cause non-nil error as well.
func TestLint_Human_MaxWarn_Error(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "human_maxwarn_err")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\n// TODO: a\n// TODO: b\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    setLintOptions(LintOptions{MaxWarn: 1})
    defer setLintOptions(LintOptions{MaxWarn: -1})
    var out bytes.Buffer
    if err := runLint(&out, dir, false, false, false); err == nil {
        t.Fatalf("expected error when warnings exceed maxWarn in human mode; out=%s", out.String())
    }
}
