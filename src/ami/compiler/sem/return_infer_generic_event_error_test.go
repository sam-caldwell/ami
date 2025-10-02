package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

// Declared generic return with type variable unifies with concrete local value.
func testReturnTypes_GenericEvent_VarParam_Unifies(t *testing.T) {
	code := "package app\nfunc Ret() (Event<T>) { var e Event<int>; return e }\n"
	f := (&source.FileSet{}).AddFile("ret_ev_decl.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeReturnTypes(af)
	for _, d := range ds {
		if d.Code == "E_RETURN_TYPE_MISMATCH" {
			t.Fatalf("unexpected mismatch: %+v", ds)
		}
	}
}

func testReturnTypes_GenericError_VarParam_Unifies(t *testing.T) {
	code := "package app\nfunc Ret() (Error<E>) { var e Error<string>; return e }\n"
	f := (&source.FileSet{}).AddFile("ret_er_decl.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeReturnTypes(af)
	for _, d := range ds {
		if d.Code == "E_RETURN_TYPE_MISMATCH" {
			t.Fatalf("unexpected mismatch: %+v", ds)
		}
	}
}
