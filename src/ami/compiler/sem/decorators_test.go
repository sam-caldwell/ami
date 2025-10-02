package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

func testAnalyzeDecorators_BuiltinsAndResolution(t *testing.T) {
	src := "package app\n@deprecated(\"msg\")\n@metrics\n@Helper\nfunc F(){}\nfunc Helper(){}\n"
	f := &source.File{Name: "d.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeDecorators(af)
	// Built-ins emit warnings (deprecated). Ensure resolution has no errors.
	for _, d := range ds {
		if d.Level == "error" {
			t.Fatalf("unexpected error diag: %+v", d)
		}
	}
}

func testAnalyzeDecorators_Unresolved_And_Conflict(t *testing.T) {
	src := "package app\n@unknown\n@dec(1)\n@dec(2)\nfunc F(){}\n"
	f := &source.File{Name: "d2.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeDecorators(af)
	var unresolved, conflict bool
	for _, d := range ds {
		if d.Code == "E_DECORATOR_UNDEFINED" {
			unresolved = true
		}
		if d.Code == "E_DECORATOR_CONFLICT" {
			conflict = true
		}
	}
	if !unresolved || !conflict {
		t.Fatalf("expected unresolved+conflict: %+v", ds)
	}
}

func testAnalyzeDecorators_Deprecated_Warning(t *testing.T) {
	src := "package app\n@deprecated(\"use G\")\nfunc F(){}\n"
	f := &source.File{Name: "d3.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeDecorators(af)
	found := false
	for _, d := range ds {
		if d.Code == "W_DEPRECATED" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected W_DEPRECATED: %+v", ds)
	}
}

func testAnalyzeDecorators_Disabled(t *testing.T) {
	defer SetDisabledDecorators() // reset
	SetDisabledDecorators("metrics")
	src := "package app\n@metrics\nfunc F(){}\n"
	f := &source.File{Name: "d4.ami", Content: src}
	p := parser.New(f)
	af, _ := p.ParseFile()
	ds := AnalyzeDecorators(af)
	found := false
	for _, d := range ds {
		if d.Code == "E_DECORATOR_DISABLED" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected E_DECORATOR_DISABLED: %+v", ds)
	}
}
