package driver

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCompile_DebugArtifacts_Deterministic_AST_IR_ASM(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("d.ami", "package app\nfunc F(){ return }\npipeline P(){ ingress; egress }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    // first run
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    ast1 := mustRead(t, filepath.Join("build", "debug", "ast", "app", "d.ast.json"))
    ir1 := mustRead(t, filepath.Join("build", "debug", "ir", "app", "d.ir.json"))
    asm1 := mustRead(t, filepath.Join("build", "debug", "asm", "app", "d.s"))
    // second run
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    ast2 := mustRead(t, filepath.Join("build", "debug", "ast", "app", "d.ast.json"))
    ir2 := mustRead(t, filepath.Join("build", "debug", "ir", "app", "d.ir.json"))
    asm2 := mustRead(t, filepath.Join("build", "debug", "asm", "app", "d.s"))
    if string(ast1) != string(ast2) { t.Fatalf("AST not deterministic") }
    if string(ir1) != string(ir2) { t.Fatalf("IR not deterministic") }
    if string(asm1) != string(asm2) { t.Fatalf("ASM not deterministic") }
}

func mustRead(t *testing.T, p string) []byte {
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read %s: %v", p, err) }
    return b
}

