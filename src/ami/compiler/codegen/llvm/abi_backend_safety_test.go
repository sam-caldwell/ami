package llvm

import (
	"strings"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func testLowerFunction_ABI_NoPtrInPublicSignature(t *testing.T) {
	// Function with param/result types that would map to ptr via mapType, but must be i64 in public ABI.
	f := ir.Function{
		Name:    "F",
		Params:  []ir.Value{{ID: "a0", Type: "string"}, {ID: "a1", Type: "Event<int>"}},
		Results: []ir.Value{{ID: "r0", Type: "map<string,int>"}},
		Blocks:  []ir.Block{{Name: "entry", Instr: []ir.Instruction{ir.Return{}}}},
	}
	s, err := lowerFunction(f)
	if err != nil {
		t.Fatalf("lower: %v", err)
	}
	// Signature should contain i64 for these ABI-handled types, and not contain "ptr %" patterns.
	if !strings.Contains(s, "define i64 @F(i64 %a0, i64 %a1)") {
		t.Fatalf("unexpected signature lowering:\n%s", s)
	}
	if strings.Contains(s, "ptr %") {
		t.Fatalf("public ABI leaked raw pointer in signature:\n%s", s)
	}
}

func testLowerFunction_RejectsRawPointerParam(t *testing.T) {
	f := ir.Function{
		Name:   "F",
		Params: []ir.Value{{ID: "p", Type: "*int"}},
		Blocks: []ir.Block{{Name: "entry", Instr: []ir.Instruction{ir.Return{}}}},
	}
	if _, err := lowerFunction(f); err == nil {
		t.Fatalf("expected error for unsafe pointer param")
	}
}

func testLowerFunction_RejectsRawPointerResult(t *testing.T) {
	f := ir.Function{
		Name:    "F",
		Results: []ir.Value{{ID: "r0", Type: "*int"}},
		Blocks:  []ir.Block{{Name: "entry", Instr: []ir.Instruction{ir.Return{}}}},
	}
	if _, err := lowerFunction(f); err == nil {
		t.Fatalf("expected error for unsafe pointer result")
	}
}

func testLowerExpr_Call_ABI_UserVsRuntimeReturn(t *testing.T) {
	// Build a module with two calls: user function (should use i64) and runtime alloc (should use ptr)
	m := ir.Module{Package: "p"}
	b := ir.Block{Name: "entry"}
	b.Instr = append(b.Instr, ir.Expr{Op: "call", Callee: "Foo", Args: []ir.Value{}, Result: &ir.Value{ID: "r0", Type: "string"}})
	b.Instr = append(b.Instr, ir.Expr{Op: "call", Callee: "ami_rt_alloc", Args: []ir.Value{{ID: "sz", Type: "int64"}}, Result: &ir.Value{ID: "p0", Type: "ptr"}})
	b.Instr = append(b.Instr, ir.Return{})
	f := ir.Function{Name: "Main", Blocks: []ir.Block{b}}
	m.Functions = []ir.Function{f}
	out, err := EmitModuleLLVM(m)
	if err != nil {
		t.Fatalf("emit: %v", err)
	}
	if !strings.Contains(out, "call i64 @Foo(") {
		t.Fatalf("user call did not use i64 return:\n%s", out)
	}
	if !strings.Contains(out, "call ptr @ami_rt_alloc(") {
		t.Fatalf("runtime alloc call did not use ptr return:\n%s", out)
	}
}
