package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify combinations of strict + failfast + max-warn in both JSON and human modes.
func TestLint_Strict_Failfast_MaxWarn_Combinations(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "modes_combo")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\n// TODO: one\n// TODO: two\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    // JSON, non-strict, max-warn=1 → error
    setLintOptions(LintOptions{MaxWarn: 1, FailFast: false})
    var out bytes.Buffer
    if err := runLint(&out, dir, true, false, false); err == nil {
        t.Fatalf("expected error with max-warn in JSON mode; out=%s", out.String())
    }

    // Human, strict, no max-warn: promote warnings to errors → error
    setLintOptions(LintOptions{MaxWarn: -1, FailFast: false})
    out.Reset()
    if err := runLint(&out, dir, false, false, true); err == nil {
        t.Fatalf("expected error in strict mode in human mode")
    }

    // JSON, strict + failfast: errors exist due to strict; should error regardless of failfast
    setLintOptions(LintOptions{MaxWarn: -1, FailFast: true})
    out.Reset()
    if err := runLint(&out, dir, true, false, true); err == nil {
        t.Fatalf("expected error in strict JSON mode")
    }

    // Reset options
    setLintOptions(LintOptions{})
}

