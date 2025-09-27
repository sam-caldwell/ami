package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Elevate W_IMPORT_ORDER to error via workspace config; expect human run to return error.
func TestLint_RuleElevation_Human_ImportOrder(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "rule_elevate_human")
    srcDir := filepath.Join(dir, "src")
    // Create workspace structure
    if err := os.MkdirAll(filepath.Join(srcDir, "a"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(filepath.Join(srcDir, "z"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte("package main\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    // Make imports intentionally unsorted to trigger W_IMPORT_ORDER
    ws.Packages[0].Package.Import = []string{"./src/z", "./src/a"}
    // Elevate to error
    if ws.Toolchain.Linter.Rules == nil { ws.Toolchain.Linter.Rules = map[string]string{} }
    ws.Toolchain.Linter.Rules["W_IMPORT_ORDER"] = "error"
    // Disable strict to isolate elevation behavior
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var out bytes.Buffer
    if err := runLint(&out, dir, false, false, false); err == nil {
        t.Fatalf("expected error due to elevated W_IMPORT_ORDER; out=%s", out.String())
    }
}

