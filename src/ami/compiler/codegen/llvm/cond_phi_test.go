package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestEmitter_CondBr_And_Phi(t *testing.T) {
    // Build a function with: entry -> condbr -> then/else -> join
    // At join, pick via phi; return value (void signature so just ensure emission)
    cond := ir.Value{ID: "c", Type: "bool"}
    vthen := ir.Value{ID: "vt", Type: "int"}
   velse := ir.Value{ID: "vf", Type: "int"}
    vout := ir.Value{ID: "vo", Type: "int"}

    b0 := ir.Block{Name: "entry", Instr: []ir.Instruction{
        ir.CondBr{Cond: cond, TrueLabel: "then", FalseLabel: "else"},
    }}
    b1 := ir.Block{Name: "then", Instr: []ir.Instruction{
        ir.Expr{Op: "lit:1", Result: &vthen},
        ir.Goto{Label: "join"},
    }}
    b2 := ir.Block{Name: "else", Instr: []ir.Instruction{
        ir.Expr{Op: "lit:2", Result: &velse},
        ir.Goto{Label: "join"},
    }}
    b3 := ir.Block{Name: "join", Instr: []ir.Instruction{
        ir.Phi{Result: vout, Incomings: []ir.PhiIncoming{{Value: vthen, Label: "then"}, {Value: velse, Label: "else"}}},
        ir.Return{},
    }}
    fn := ir.Function{Name: "F", Blocks: []ir.Block{b0,b1,b2,b3}}
    m := ir.Module{Package: "app", Functions: []ir.Function{fn}}

    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "entry:\n  br i1 %c, label %then, label %else\n") {
        t.Fatalf("missing condbr:\n%s", out)
    }
    if !strings.Contains(out, "join:\n  %vo = phi i64 [ %vt, %then ], [ %vf, %else ]\n") {
        t.Fatalf("missing phi at join:\n%s", out)
    }
}

