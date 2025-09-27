package main

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_Verbose_WritesASTAndIRDebug(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "verbose_debug")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Workspace and a simple .ami
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    src := "package app\nfunc F(){}\npipeline P() { ingress; egress }\n"
    if err := os.WriteFile(filepath.Join(dir, "src", "main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    if err := runBuild(os.Stdout, dir, false, true); err != nil { t.Fatalf("runBuild: %v", err) }
    // Expect AST JSON under build/debug/ast/app
    astDir := filepath.Join(dir, "build", "debug", "ast", ws.Packages[0].Package.Name)
    matches, _ := filepath.Glob(filepath.Join(astDir, "*.ast.json"))
    if len(matches) == 0 { t.Fatalf("expected AST debug files in %s", astDir) }
    // Expect IR JSON under build/debug/ir/app
    irDir := filepath.Join(dir, "build", "debug", "ir", ws.Packages[0].Package.Name)
    irMatches, _ := filepath.Glob(filepath.Join(irDir, "*.ir.json"))
    if len(irMatches) == 0 { t.Fatalf("expected IR debug files in %s", irDir) }
}

