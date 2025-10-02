package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

// Mixed: single multi-result call plus scalar; expect arity diag for generic Owned and
// a type mismatch for the scalar vs declared Error<E>.
func testReturnTypesWithSigs_Mixed_ExpandPlusScalar(t *testing.T) {
	code := "package app\n" +
		"func Producer() (Owned<int,string>, Error<int>) { return }\n" +
		"func F() (Owned<T>, Error<E>) { return Producer(), 1 }\n"
	var fs source.FileSet
	f := fs.AddFile("u_mixed1.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFileCollect()
	results := map[string][]string{
		"Producer": {"Owned<int,string>", "Error<int>"},
	}
	ds := AnalyzeReturnTypesWithSigs(af, results)
	var haveArityIdx0 bool
	var haveMismatchIdx1 bool
	for _, d := range ds {
		if d.Data == nil {
			continue
		}
		var idx int
		if v, ok := d.Data["index"].(int); ok {
			idx = v
		} else if vf, ok := d.Data["index"].(float64); ok {
			idx = int(vf)
		}
		switch d.Code {
		case "E_GENERIC_ARITY_MISMATCH":
			if idx == 0 {
				haveArityIdx0 = true
			}
		case "E_CALL_RESULT_MISMATCH":
			if idx == 1 {
				haveMismatchIdx1 = true
			}
		}
	}
	if !haveArityIdx0 || !haveMismatchIdx1 {
		t.Fatalf("expected arity idx0 and mismatch idx1; got %+v", ds)
	}
}

// Two calls with an interleaved scalar: only the Error<E> vs Error<string,string> should
// produce a generic arity mismatch at index 2.
func testReturnTypesWithSigs_Mixed_TwoCallsAndScalar(t *testing.T) {
	code := "package app\n" +
		"func P() (Owned<string>) { return }\n" +
		"func Q() (Error<string,string>) { return }\n" +
		"func F() (Owned<T>, int, Error<E>) { return P(), 3, Q() }\n"
	var fs source.FileSet
	f := fs.AddFile("u_mixed2.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFileCollect()
	results := map[string][]string{
		"P": {"Owned<string>"},
		"Q": {"Error<string,string>"},
	}
	ds := AnalyzeReturnTypesWithSigs(af, results)
	countIdx2 := 0
	for _, d := range ds {
		if d.Code != "E_GENERIC_ARITY_MISMATCH" || d.Data == nil {
			continue
		}
		var idx int
		if v, ok := d.Data["index"].(int); ok {
			idx = v
		} else if vf, ok := d.Data["index"].(float64); ok {
			idx = int(vf)
		}
		if idx == 2 {
			countIdx2++
		}
	}
	if countIdx2 != 1 {
		t.Fatalf("expected 1 generic arity mismatch at index 2; got %d (%+v)", countIdx2, ds)
	}
}

// A multi-result call that does not fit declared arity should report a return arity mismatch.
func testReturnTypesWithSigs_MultiResult_DoesNotFitArity(t *testing.T) {
	code := "package app\n" +
		"func P() (int,int,int) { return }\n" +
		"func F() (int,int) { return P() }\n"
	var fs source.FileSet
	f := fs.AddFile("u_mixed3.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFileCollect()
	results := map[string][]string{"P": {"int", "int", "int"}}
	ds := AnalyzeReturnTypesWithSigs(af, results)
	found := false
	for _, d := range ds {
		if d.Code == "E_RETURN_TYPE_MISMATCH" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected E_RETURN_TYPE_MISMATCH for arity; got %+v", ds)
	}
}
