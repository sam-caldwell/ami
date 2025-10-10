package llvm

import (
    "strings"
    "testing"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestLowerExpr_Field_Fallback_Types(t *testing.T) {
    base := ir.Value{ID: "b", Type: "UnknownType"}
    // bool fallback
    s := lowerExpr(ir.Expr{Op: "field.a", Args: []ir.Value{base}, Result: &ir.Value{ID: "r1", Type: "bool"}})
    if !strings.Contains(s, "icmp ne i1 0, 1") { t.Fatalf("bool fallback: %s", s) }
    // int fallback
    s = lowerExpr(ir.Expr{Op: "field.a", Args: []ir.Value{base}, Result: &ir.Value{ID: "r2", Type: "int"}})
    if !strings.Contains(s, "add i64 0, 0") { t.Fatalf("int fallback: %s", s) }
    // float fallback
    s = lowerExpr(ir.Expr{Op: "field.a", Args: []ir.Value{base}, Result: &ir.Value{ID: "r3", Type: "float64"}})
    if !strings.Contains(s, "fadd double 0.0, 0.0") { t.Fatalf("float fallback: %s", s) }
    // ptr fallback
    s = lowerExpr(ir.Expr{Op: "field.a", Args: []ir.Value{base}, Result: &ir.Value{ID: "r4", Type: "ptr"}})
    if !strings.Contains(s, "getelementptr i8, ptr null") { t.Fatalf("ptr fallback: %s", s) }
}

func TestLowerExpr_Literals_IntFloatBool(t *testing.T) {
    // int literal
    s := lowerExpr(ir.Expr{Op: "lit:42", Result: &ir.Value{ID: "x", Type: "int"}})
    if !strings.Contains(s, "add i64 0, 42") { t.Fatalf("int lit: %s", s) }
    // float zero literal
    s = lowerExpr(ir.Expr{Op: "lit:0.0", Result: &ir.Value{ID: "y", Type: "float64"}})
    if !strings.Contains(s, "fadd double 0.0, 0.0") { t.Fatalf("float zero: %s", s) }
    // bool false literal
    s = lowerExpr(ir.Expr{Op: "lit:false", Result: &ir.Value{ID: "z", Type: "bool"}})
    if !strings.Contains(s, "icmp ne i1 0, 1") { t.Fatalf("bool false: %s", s) }
}

func TestLowerExpr_Frem_WithResult(t *testing.T) {
    s := lowerExpr(ir.Expr{Op: "frem", Args: []ir.Value{{ID: "a", Type: "float64"}, {ID: "b", Type: "float64"}}, Result: &ir.Value{ID: "r", Type: "float64"}})
    if !strings.Contains(s, "frem") { t.Fatalf("frem: %s", s) }
}

func TestLowerExpr_Call_IntrinsicAndImmediates(t *testing.T) {
    // math intrinsic with result
    s := lowerExpr(ir.Expr{Op: "call", Callee: "math.Sqrt", Args: []ir.Value{{ID: "x", Type: "float64"}}, Result: &ir.Value{ID: "r", Type: "float64"}})
    if !strings.Contains(s, "@llvm.sqrt.f64") { t.Fatalf("intrinsic: %s", s) }
    // call with immediate function pointer and null pointer arg
    args := []ir.Value{{ID: "#@HANDLER", Type: "ptr"}, {ID: "#null", Type: "ptr"}}
    s = lowerExpr(ir.Expr{Op: "call", Callee: "ami_rt_install_handler_thunk", Args: args})
    if !strings.Contains(s, "@HANDLER") || !strings.Contains(s, "ptr null") { t.Fatalf("immediates: %s", s) }
}

func TestLowerExpr_Call_MultiResult(t *testing.T) {
    // multi-result call aggregates then extractvalue
    res := []ir.Value{{ID: "r0", Type: "int"}, {ID: "r1", Type: "float64"}}
    s := lowerExpr(ir.Expr{Op: "call", Callee: "foo", Args: []ir.Value{{ID: "x", Type: "int"}}, Results: res})
    if !strings.Contains(s, "extractvalue") { t.Fatalf("multi-result: %s", s) }
}

func TestLowerExpr_Select(t *testing.T) {
    args := []ir.Value{{ID: "c", Type: "bool"}, {ID: "t", Type: "int"}, {ID: "f", Type: "int"}}
    s := lowerExpr(ir.Expr{Op: "select", Args: args, Result: &ir.Value{ID: "r", Type: "int"}})
    if !strings.Contains(s, "select i1 %c") { t.Fatalf("select: %s", s) }
}

func TestLowerExpr_Field_SuccessPath(t *testing.T) {
    base := ir.Value{ID: "b", Type: "Struct{a:int}"}
    s := lowerExpr(ir.Expr{Op: "field.a", Args: []ir.Value{base}, Result: &ir.Value{ID: "r", Type: "int"}})
    if !strings.Contains(s, "getelementptr") || !strings.Contains(s, "load i64") { t.Fatalf("field success: %s", s) }
}

func TestLowerExpr_EventPayload_Field_HelperCalls(t *testing.T) {
    // Ensure event.payload.field.<path> lowers to runtime helper calls by type.
    ev := ir.Value{ID: "ev", Type: "Event<int>"}
    // int
    s := lowerExpr(ir.Expr{Op: "event.payload.field.k", Args: []ir.Value{ev}, Result: &ir.Value{ID: "i", Type: "int"}})
    if !strings.Contains(s, "@ami_rt_event_get_i64") { t.Fatalf("int helper missing: %s", s) }
    // float
    s = lowerExpr(ir.Expr{Op: "event.payload.field.k", Args: []ir.Value{ev}, Result: &ir.Value{ID: "d", Type: "float64"}})
    if !strings.Contains(s, "@ami_rt_event_get_double") { t.Fatalf("double helper missing: %s", s) }
    // bool
    s = lowerExpr(ir.Expr{Op: "event.payload.field.k", Args: []ir.Value{ev}, Result: &ir.Value{ID: "b", Type: "bool"}})
    if !strings.Contains(s, "@ami_rt_event_get_bool") { t.Fatalf("bool helper missing: %s", s) }
    // string
    s = lowerExpr(ir.Expr{Op: "event.payload.field.k", Args: []ir.Value{ev}, Result: &ir.Value{ID: "s", Type: "string"}})
    if !strings.Contains(s, "@ami_rt_event_get_string") { t.Fatalf("string helper missing: %s", s) }
}
