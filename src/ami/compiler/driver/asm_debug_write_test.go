package driver

import (
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestWriteAsmDebug_EmitsListing(t *testing.T) {
    f := &ast.File{}
    // add a minimal pipeline with Collect and edge to exercise pseudo-ops
    f.Decls = append(f.Decls, &ast.PipelineDecl{Name: "P"})
    m := ir.Module{Functions: []ir.Function{{Name: "F", Blocks: []ir.Block{{Name: "entry", Instr: []ir.Instruction{ir.Var{}, ir.Return{}}}}}}}
    path, err := writeAsmDebug("main", "u1", f, m)
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    if len(b) == 0 { t.Fatalf("expected non-empty asm listing") }
}

