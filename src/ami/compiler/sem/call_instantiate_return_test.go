package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "testing"
)

// Verify we instantiate generic function return types from call-site argument types
// and propagate through return unification.
func TestReturn_CallGeneric_Instantiate_OK(t *testing.T) {
    src := `package p
func g(x Event<T>) Event<T> { }
func f(ev Event<string>) Event<string> { return g(ev) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_RETURN_TYPE_MISMATCH" || d.Code == "E_TYPE_UNINFERRED" || d.Code == "E_CALL_ARG_TYPE_MISMATCH" {
            t.Fatalf("unexpected diag: %v", d)
        }
    }
}

func TestReturn_CallGeneric_Instantiate_Mismatch_Error(t *testing.T) {
    src := `package p
func g(x Event<T>) Event<T> { }
func f(ev Event<int>) Event<string> { return g(ev) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics {
        if d.Code == "E_RETURN_TYPE_MISMATCH" {
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("expected E_RETURN_TYPE_MISMATCH; diags=%v", res.Diagnostics)
    }
}

// Container literals without explicit type args should infer element/key/value types
// and unify with declared return type.
func TestReturn_ContainerLiteral_Infer_OK(t *testing.T) {
    src := `package p
func f() slice<int> { var s = slice{1,2,3}; return s }
func g() map<string,int> { var m = map{"a":1}; return m }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_RETURN_TYPE_MISMATCH" || d.Code == "E_ASSIGN_TYPE_MISMATCH" || d.Code == "E_TYPE_MISMATCH" {
            t.Fatalf("unexpected diag: %v", d)
        }
    }
}
