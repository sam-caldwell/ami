package ir

import (
	"encoding/json"
	"testing"
)

func testEncode_ExprCall_IncludesCallee(t *testing.T) {
	temp := Value{ID: "t1", Type: "any"}
	call := Expr{Op: "call", Callee: "Foo", Args: []Value{{ID: "a", Type: "int"}}, Result: &temp}
	f := Function{Name: "F", Blocks: []Block{{Name: "entry", Instr: []Instruction{call}}}}
	m := Module{Package: "p", Functions: []Function{f}}
	b, err := EncodeModule(m)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(b, &obj); err != nil {
		t.Fatalf("json: %v", err)
	}
	// quick presence check: walk
	fns := obj["functions"].([]any)
	fn := fns[0].(map[string]any)
	blks := fn["blocks"].([]any)
	blk := blks[0].(map[string]any)
	instrs := blk["instrs"].([]any)
	in := instrs[0].(map[string]any)
	if in["op"] != "EXPR" {
		t.Fatalf("op: %v", in["op"])
	}
	expr := in["expr"].(map[string]any)
	if expr["op"] != "call" || expr["callee"] != "Foo" {
		t.Fatalf("expr: %v", expr)
	}
}

func testEncode_ExprCall_ArgsIncludeTypes(t *testing.T) {
	temp := Value{ID: "t1", Type: "int"}
	call := Expr{Op: "call", Callee: "Foo", Args: []Value{{ID: "a", Type: "string"}}, Result: &temp}
	f := Function{Name: "F", Blocks: []Block{{Name: "entry", Instr: []Instruction{call}}}}
	m := Module{Package: "p", Functions: []Function{f}}
	b, err := EncodeModule(m)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(b, &obj); err != nil {
		t.Fatalf("json: %v", err)
	}
	fns := obj["functions"].([]any)
	fn := fns[0].(map[string]any)
	blks := fn["blocks"].([]any)
	blk := blks[0].(map[string]any)
	instrs := blk["instrs"].([]any)
	in := instrs[0].(map[string]any)
	expr := in["expr"].(map[string]any)
	args := expr["args"].([]any)
	a0 := args[0].(map[string]any)
	if a0["type"] != "string" {
		t.Fatalf("arg type: %v", a0["type"])
	}
}

func testEncode_ExprCall_SignatureTypes(t *testing.T) {
	temp := Value{ID: "t1", Type: "int"}
	call := Expr{Op: "call", Callee: "Foo", Args: []Value{{ID: "a", Type: "string"}, {ID: "b", Type: "int"}}, Result: &temp}
	f := Function{Name: "F", Blocks: []Block{{Name: "entry", Instr: []Instruction{call}}}}
	m := Module{Package: "p", Functions: []Function{f}}
	b, err := EncodeModule(m)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(b, &obj); err != nil {
		t.Fatalf("json: %v", err)
	}
	fns := obj["functions"].([]any)
	fn := fns[0].(map[string]any)
	blks := fn["blocks"].([]any)
	blk := blks[0].(map[string]any)
	instrs := blk["instrs"].([]any)
	in := instrs[0].(map[string]any)
	expr := in["expr"].(map[string]any)
	at := expr["argTypes"].([]any)
	rt := expr["retTypes"].([]any)
	if len(at) != 2 || at[0] != "string" || at[1] != "int" {
		t.Fatalf("argTypes: %+v", at)
	}
	if len(rt) != 1 || rt[0] != "int" {
		t.Fatalf("retTypes: %+v", rt)
	}
}

func testEncode_ExprCall_MultiResult_EncodesResultsAndRetTypes(t *testing.T) {
	// Construct a call expression that yields two results and ensure both the
	// JSON-level `results` field and the derived `retTypes` array reflect them.
	r0 := Value{ID: "t1", Type: "int"}
	r1 := Value{ID: "t2", Type: "string"}
	call := Expr{Op: "call", Callee: "Pair", Args: []Value{{ID: "a", Type: "bool"}}, Results: []Value{r0, r1}}
	f := Function{Name: "F", Blocks: []Block{{Name: "entry", Instr: []Instruction{call}}}}
	m := Module{Package: "p", Functions: []Function{f}}
	b, err := EncodeModule(m)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(b, &obj); err != nil {
		t.Fatalf("json: %v", err)
	}
	fns := obj["functions"].([]any)
	fn := fns[0].(map[string]any)
	blks := fn["blocks"].([]any)
	blk := blks[0].(map[string]any)
	instrs := blk["instrs"].([]any)
	in := instrs[0].(map[string]any)
	expr := in["expr"].(map[string]any)
	// Verify results array is present with both entries
	rsAny, ok := expr["results"].([]any)
	if !ok || len(rsAny) != 2 {
		t.Fatalf("results missing or wrong len: %#v", expr["results"])
	}
	if rsAny[0].(map[string]any)["type"] != "int" || rsAny[1].(map[string]any)["type"] != "string" {
		t.Fatalf("result types mismatch: %v", rsAny)
	}
	// Verify retTypes mirrors results list
	rt := expr["retTypes"].([]any)
	if len(rt) != 2 || rt[0] != "int" || rt[1] != "string" {
		t.Fatalf("retTypes: %+v", rt)
	}
	// Verify sig.results mirrors result types
	sig := expr["sig"].(map[string]any)
	sigRs := sig["results"].([]any)
	if len(sigRs) != 2 || sigRs[0] != "int" || sigRs[1] != "string" {
		t.Fatalf("sig.results: %+v", sigRs)
	}
}
