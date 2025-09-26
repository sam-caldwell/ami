package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "testing"
)

func TestReturn_Tuple_CallInstantiate_OK(t *testing.T) {
    src := `package p
func h(x Event<T>) (Event<T>, Error<T>) { }
func f(ev Event<string>) (Event<string>, Error<string>) { return h(ev) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_RETURN_TYPE_MISMATCH" || d.Code == "E_TYPE_UNINFERRED" || d.Code == "E_CALL_ARG_TYPE_MISMATCH" {
            t.Fatalf("unexpected diag: %v", d)
        }
    }
}

func TestReturn_Tuple_CallInstantiate_TypeMismatch_Error(t *testing.T) {
    src := `package p
func h(x Event<T>) (Event<T>, Error<T>) { }
func f(ev Event<int>) (Event<string>, Error<string>) { return h(ev) }`
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

