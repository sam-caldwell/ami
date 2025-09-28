package llvm

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// TestEmitModuleLLVMForTarget_SetsTriple_And_CollectsExterns exercises SetTargetTriple,
// RequireExtern, lowerVar, and lowerAssign paths by emitting a small module that
// references panic/alloc and includes VAR/ASSIGN instructions.
func TestEmitModuleLLVMForTarget_SetsTriple_And_CollectsExterns(t *testing.T) {
    m := ir.Module{Package: "p"}
    // Build a simple function with a var, assign, and expressions that trigger externs.
    f := ir.Function{Name: "F", Results: nil}
    b := ir.Block{Name: "entry"}
    // var x : int
    b.Instr = append(b.Instr, ir.Var{Name: "x", Type: "int", Result: ir.Value{ID: "x0", Type: "int"}})
    // assign x = c1
    b.Instr = append(b.Instr, ir.Assign{DestID: "x0", Src: ir.Value{ID: "c1", Type: "int"}})
    // expressions that imply extern requirements
    b.Instr = append(b.Instr, ir.Expr{Op: "panic", Args: []ir.Value{{ID: "code", Type: "int32"}}})
    b.Instr = append(b.Instr, ir.Expr{Op: "call", Callee: "ami_rt_alloc", Args: []ir.Value{{ID: "sz", Type: "int64"}}})
    // return
    b.Instr = append(b.Instr, ir.Return{})
    f.Blocks = []ir.Block{b}
    m.Functions = []ir.Function{f}

    out, err := EmitModuleLLVMForTarget(m, "x86_64-unknown-linux-gnu")
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    // target triple must be set
    if !strings.Contains(out, "target triple = \"x86_64-unknown-linux-gnu\"") {
        t.Fatalf("expected target triple set; got:\n%s", out)
    }
    // extern declarations present
    if !strings.Contains(out, "declare void @ami_rt_panic(i32)") {
        t.Fatalf("expected panic extern; got:\n%s", out)
    }
    if !strings.Contains(out, "declare ptr @ami_rt_alloc(i64)") {
        t.Fatalf("expected alloc extern; got:\n%s", out)
    }
    // var/assign scaffolds appear as deterministic comments
    if !strings.Contains(out, "; var x : i64 as %x0") { // int → i64
        t.Fatalf("expected var lowering comment; got:\n%s", out)
    }
    if !strings.Contains(out, "; assign %x0 = %c1") {
        t.Fatalf("expected assign lowering comment; got:\n%s", out)
    }
}

