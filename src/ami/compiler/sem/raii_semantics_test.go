package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestRAII_OwnedParam_NotReleased_Error(t *testing.T) {
	src := `package p
func f(ctx Context, ev Event<string>, r Owned<string>, st State) Event<string> { }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_RAII_OWNED_NOT_RELEASED" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_RAII_OWNED_NOT_RELEASED; diags=%v", res.Diagnostics)
	}
}

func TestRAII_OwnedParam_Release_OK(t *testing.T) {
	src := `package p
func f(ctx Context, ev Event<string>, r Owned<string>, st State) Event<string> { mutate(r.Close()) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_RAII_OWNED_NOT_RELEASED" {
			t.Fatalf("unexpected not-released: %v", d)
		}
	}
}

// Transfer not asserted here due to token-level analysis; method-style release suffices for scaffold.

func TestRAII_DoubleRelease_Error(t *testing.T) {
	src := `package p
func release(o Owned<string>) {}
func f(ctx Context, ev Event<string>, r Owned<string>, st State) Event<string> { mutate(release(r)); mutate(release(r)) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_RAII_DOUBLE_RELEASE" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_RAII_DOUBLE_RELEASE; diags=%v", res.Diagnostics)
	}
}

func TestRAII_UseAfterRelease_Error(t *testing.T) {
	src := `package p
func release(o Owned<string>) {}
func f(ctx Context, ev Event<string>, r Owned<string>, st State) Event<string> { mutate(release(r)) r }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_RAII_USE_AFTER_RELEASE" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_RAII_USE_AFTER_RELEASE; diags=%v", res.Diagnostics)
	}
}
