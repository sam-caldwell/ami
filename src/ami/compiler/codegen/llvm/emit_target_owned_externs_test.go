package llvm

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Ensure EmitModuleLLVMForTarget collects Owned-related externs from both Expr and Defer contexts.
func TestEmitModuleLLVMForTarget_CollectsOwnedExterns(t *testing.T) {
    m := ir.Module{Package: "p"}
    f := ir.Function{Name: "F"}
    b := ir.Block{Name: "entry"}
    // Explicit calls to owned helpers in expressions
    b.Instr = append(b.Instr,
        ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{{ID: "h", Type: "ptr"}}, Result: &ir.Value{ID: "l", Type: "int64"}},
        ir.Expr{Op: "call", Callee: "ami_rt_owned_ptr", Args: []ir.Value{{ID: "h", Type: "ptr"}}, Result: &ir.Value{ID: "p", Type: "ptr"}},
        ir.Expr{Op: "call", Callee: "ami_rt_owned_new", Args: []ir.Value{{ID: "p", Type: "ptr"}, {ID: "l", Type: "int64"}}, Result: &ir.Value{ID: "h2", Type: "ptr"}},
    )
    // Defer zeroize_owned should also trigger extern
    b.Instr = append(b.Instr, ir.Defer{Expr: ir.Expr{Op: "call", Callee: "ami_rt_zeroize_owned", Args: []ir.Value{{ID: "h2", Type: "ptr"}}}})
    // Return
    b.Instr = append(b.Instr, ir.Return{})
    f.Blocks = []ir.Block{b}
    m.Functions = []ir.Function{f}

    out, err := EmitModuleLLVMForTarget(m, "x86_64-unknown-linux-gnu")
    if err != nil { t.Fatalf("emit: %v", err) }
    // Validate externs
    must := []string{
        "declare i64 @ami_rt_owned_len(ptr)",
        "declare ptr @ami_rt_owned_ptr(ptr)",
        "declare ptr @ami_rt_owned_new(i8*, i64)",
        "declare void @ami_rt_zeroize_owned(ptr)",
    }
    for _, s := range must {
        if !strings.Contains(out, s) { t.Fatalf("missing extern %q in:\n%s", s, out) }
    }
}

