package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
    "strings"
    "testing"
)

func TestAnalyzeFile_FunctionTypeInference(t *testing.T) {
	src := `package p
import "util"
func f(ev Event<string>) (Event<string>, error) { }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	if len(res.Diagnostics) != 0 {
		t.Fatalf("unexpected diags: %+v", res.Diagnostics)
	}
	// scope lookup for imported package name
	if obj := res.Scope.Lookup("util"); obj == nil || obj.Type.String() != "package" {
		t.Fatalf("import not in scope: %+v", obj)
	}
	// function type constructed
	obj := res.Scope.Lookup("f")
	if obj == nil {
		t.Fatalf("function f not in scope")
	}
	fn, ok := obj.Type.(types.Function)
	if !ok {
		t.Fatalf("wrong type: %T", obj.Type)
	}
    if len(fn.Params) != 1 || len(fn.Results) != 2 {
        t.Fatalf("wrong arity: %+v", fn)
    }
	// check a few renderings
    if fn.Params[0].String() != "Event<string>" {
        t.Fatalf("param1: %s", fn.Params[1].String())
    }
    if fn.Results[0].String() != "Event<string>" || strings.ToLower(fn.Results[1].String()) != "error" {
        t.Fatalf("result: %s", fn.Results[0].String())
    }
}
