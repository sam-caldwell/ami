package llvm

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestEmitLLVM_SimpleFunctionSignatureAndReturn(t *testing.T) {
    m := ir.Module{Package: "app", Functions: []ir.Function{{
        Name:    "Foo",
        Params:  []ir.Value{{ID: "x", Type: "int"}},
        Results: []ir.Value{{Type: "int"}},
        Blocks:  []ir.Block{{Name: "entry", Instr: []ir.Instruction{ir.Return{Values: []ir.Value{{ID: "x", Type: "int"}}}}}},
    }}}
    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    // Header
    if !strings.Contains(out, "target triple = \"arm64-apple-macosx\"") {
        t.Fatalf("missing target triple: %s", out)
    }
    // Externs are added based on usage; simple function need not declare any.
    // Signature
    if !strings.Contains(out, "define i64 @Foo(i64 %x) {") {
        t.Fatalf("signature mismatch: %s", out)
    }
    // Return
    if !strings.Contains(out, "  ret i64 %x\n") {
        t.Fatalf("return mismatch: %s", out)
    }
}

func TestEmitLLVM_ExprCallWithResult(t *testing.T) {
    f := ir.Function{
        Name:    "Main",
        Params:  []ir.Value{{ID: "x", Type: "int"}},
        Results: []ir.Value{{Type: "int"}},
        Blocks: []ir.Block{{Name: "entry", Instr: []ir.Instruction{
            ir.Expr{Op: "call", Callee: "Foo", Args: []ir.Value{{ID: "x", Type: "int"}}, Result: &ir.Value{ID: "t1", Type: "int"}},
            ir.Return{Values: []ir.Value{{ID: "t1", Type: "int"}}},
        }}},
    }
    out, err := EmitModuleLLVM(ir.Module{Package: "app", Functions: []ir.Function{f}})
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "%t1 = call i64 @Foo(i64 %x)") {
        t.Fatalf("call not lowered as expected: %s", out)
    }
    if !strings.Contains(out, "  ret i64 %t1\n") {
        t.Fatalf("return not lowered as expected: %s", out)
    }
}

func TestEmitLLVM_MultiResultFunctionAndReturn(t *testing.T) {
    // define {i64,i64} @Pair()
    f := ir.Function{Name: "Pair", Results: []ir.Value{{Type: "int"}, {Type: "int"}}, Blocks: []ir.Block{{Name: "entry", Instr: []ir.Instruction{
        ir.Return{Values: []ir.Value{{ID: "a", Type: "int"}, {ID: "b", Type: "int"}}},
    }}}}
    out, err := lowerFunction(f)
    if err != nil { t.Fatalf("lower: %v", err) }
    if !strings.Contains(out, "define {i64, i64} @Pair(") {
        t.Fatalf("multi-result signature missing: %s", out)
    }
    if !strings.Contains(out, "insertvalue {i64, i64} undef, i64 %a, 0") {
        t.Fatalf("insertvalue chain missing: %s", out)
    }
}

func TestEmitLLVM_MultiBlockAndUnknownOp(t *testing.T) {
    f := ir.Function{
        Name:    "G",
        Params:  nil,
        Results: []ir.Value{{Type: "int"}},
        Blocks: []ir.Block{
            {Name: "entry", Instr: []ir.Instruction{ir.Expr{Op: "mystery", Args: nil, Result: &ir.Value{ID: "t1", Type: "int"}}}},
            {Name: "b2", Instr: []ir.Instruction{ir.Return{Values: []ir.Value{{ID: "t1", Type: "int"}}}}},
        },
    }
    out, err := EmitModuleLLVM(ir.Module{Package: "app", Functions: []ir.Function{f}})
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "entry:") || !strings.Contains(out, "b2:") {
        t.Fatalf("missing block labels: %s", out)
    }
    if !strings.Contains(out, "; expr mystery") {
        t.Fatalf("unknown op not emitted as comment: %s", out)
    }
}

func TestEmitLLVM_ComparisonsAndIntLiteral(t *testing.T) {
    // %t1 = lit:42 ; ret t1
    f1 := ir.Function{ Name: "L", Results: []ir.Value{{Type: "int"}}, Blocks: []ir.Block{{Name: "entry", Instr: []ir.Instruction{
        ir.Expr{Op: "lit:42", Result: &ir.Value{ID: "t1", Type: "int"}},
        ir.Return{Values: []ir.Value{{ID: "t1", Type: "int"}}},
    }}}}
    // %t2 = icmp eq i64 %x, %y
    f2 := ir.Function{ Name: "C", Params: []ir.Value{{ID: "x", Type: "int"}, {ID: "y", Type: "int"}}, Results: []ir.Value{{Type: "bool"}}, Blocks: []ir.Block{{Name: "entry", Instr: []ir.Instruction{
        ir.Expr{Op: "eq", Args: []ir.Value{{ID: "x", Type: "int"}, {ID: "y", Type: "int"}}, Result: &ir.Value{ID: "t2", Type: "bool"}},
        ir.Return{Values: []ir.Value{{ID: "t2", Type: "bool"}}},
    }}}}
    out, err := EmitModuleLLVM(ir.Module{Package: "app", Functions: []ir.Function{f1, f2}})
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "%t1 = add i64 0, 42") { t.Fatalf("literal not lowered: %s", out) }
    if !strings.Contains(out, "%t2 = icmp eq i64 %x, %y") { t.Fatalf("icmp not lowered: %s", out) }
}

func TestEmitLLVM_ModAndLogicalAnd(t *testing.T) {
    // t1 = mod(x, y) ; t2 = and(b1,b2)
    f := ir.Function{
        Name:    "Ops",
        Params:  []ir.Value{{ID: "x", Type: "int"}, {ID: "y", Type: "int"}, {ID: "b1", Type: "bool"}, {ID: "b2", Type: "bool"}},
        Results: []ir.Value{},
        Blocks: []ir.Block{{Name: "entry", Instr: []ir.Instruction{
            ir.Expr{Op: "mod", Args: []ir.Value{{ID: "x", Type: "int"}, {ID: "y", Type: "int"}}, Result: &ir.Value{ID: "t1", Type: "int"}},
            ir.Expr{Op: "and", Args: []ir.Value{{ID: "b1", Type: "bool"}, {ID: "b2", Type: "bool"}}, Result: &ir.Value{ID: "t2", Type: "bool"}},
            ir.Return{},
        }}},
    }
    out, err := EmitModuleLLVM(ir.Module{Package: "app", Functions: []ir.Function{f}})
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "%t1 = srem i64 %x, %y") {
        t.Fatalf("mod not lowered as srem: %s", out)
    }
    if !strings.Contains(out, "%t2 = and i1 %b1, %b2") {
        t.Fatalf("and not lowered for bools: %s", out)
    }
}

func TestEmitLLVM_Shifts_Neg_Xor_Bnot(t *testing.T) {
    f := ir.Function{
        Name:   "S",
        Params: []ir.Value{{ID: "x", Type: "int"}, {ID: "y", Type: "int"}},
        Blocks: []ir.Block{{Name: "entry", Instr: []ir.Instruction{
            ir.Expr{Op: "shl", Args: []ir.Value{{ID: "x", Type: "int"}, {ID: "y", Type: "int"}}, Result: &ir.Value{ID: "t1", Type: "int"}},
            ir.Expr{Op: "shr", Args: []ir.Value{{ID: "x", Type: "int"}, {ID: "y", Type: "int"}}, Result: &ir.Value{ID: "t2", Type: "int"}},
            ir.Expr{Op: "xor", Args: []ir.Value{{ID: "x", Type: "int"}, {ID: "y", Type: "int"}}, Result: &ir.Value{ID: "t3", Type: "int"}},
            ir.Expr{Op: "bor", Args: []ir.Value{{ID: "x", Type: "int"}, {ID: "y", Type: "int"}}, Result: &ir.Value{ID: "t6", Type: "int"}},
            ir.Expr{Op: "neg", Args: []ir.Value{{ID: "x", Type: "int"}}, Result: &ir.Value{ID: "t4", Type: "int"}},
            ir.Expr{Op: "bnot", Args: []ir.Value{{ID: "x", Type: "int"}}, Result: &ir.Value{ID: "t5", Type: "int"}},
            ir.Return{},
        }}},
    }
    out, err := EmitModuleLLVM(ir.Module{Package: "app", Functions: []ir.Function{f}})
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "%t1 = shl i64 %x, %y") { t.Fatalf("shl not lowered: %s", out) }
    if !strings.Contains(out, "%t2 = ashr i64 %x, %y") { t.Fatalf("shr not lowered: %s", out) }
    if !strings.Contains(out, "%t3 = xor i64 %x, %y") { t.Fatalf("xor not lowered: %s", out) }
    if !strings.Contains(out, "%t4 = sub i64 0, %x") { t.Fatalf("neg not lowered: %s", out) }
    if !strings.Contains(out, "%t5 = xor i64 %x, -1") { t.Fatalf("bnot not lowered: %s", out) }
    if !strings.Contains(out, "%t6 = or i64 %x, %y") { t.Fatalf("bor not lowered: %s", out) }
}

func TestEmitLLVM_NotBool(t *testing.T) {
    f := ir.Function{
        Name:   "N",
        Params: []ir.Value{{ID: "b", Type: "bool"}},
        Blocks: []ir.Block{{Name: "entry", Instr: []ir.Instruction{
            ir.Expr{Op: "not", Args: []ir.Value{{ID: "b", Type: "bool"}}, Result: &ir.Value{ID: "t", Type: "bool"}},
            ir.Return{},
        }}},
    }
    out, err := EmitModuleLLVM(ir.Module{Package: "app", Functions: []ir.Function{f}})
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "%t = xor i1 %b, true") { t.Fatalf("not not lowered: %s", out) }
}
