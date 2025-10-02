package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"testing"
)

func testRAII_Leak_OnOwnedResult_NotReleased(t *testing.T) {
	// Function G returns Owned (generic suppressed in parser); assigning to x marks x as owned.
	// Since there is no release/transfer, analyzer should emit E_RAII_LEAK.
	code := "package app\nfunc G() (Owned) { return }\nfunc F(){ var x; x = G() }\n"
	f := (&source.FileSet{}).AddFile("leak.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFileCollect()
	ds := AnalyzeRAII(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_RAII_LEAK" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_RAII_LEAK, got: %+v", ds)
	}
}

func testRAII_Transfer_Unowned_Error(t *testing.T) {
	// H takes Owned parameter; passing y (not owned) should emit E_RAII_TRANSFER_UNOWNED.
	code := "package app\nfunc H(a Owned){}\nfunc F(){ var y; H(y) }\n"
	f := (&source.FileSet{}).AddFile("xfer.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFileCollect()
	ds := AnalyzeRAII(af)
	has := false
	for _, d := range ds {
		if d.Code == "E_RAII_TRANSFER_UNOWNED" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_RAII_TRANSFER_UNOWNED, got: %+v", ds)
	}
}

func testRAII_NoLeak_WhenReleased(t *testing.T) {
	code := "package app\nfunc G() (Owned) { return }\nfunc F(){ var r; r = G(); release(r) }\n"
	f := (&source.FileSet{}).AddFile("noleak.ami", code)
	p := parser.New(f)
	af, _ := p.ParseFileCollect()
	ds := AnalyzeRAII(af)
	for _, d := range ds {
		if d.Code == "E_RAII_LEAK" {
			t.Fatalf("unexpected leak: %+v", ds)
		}
	}
}
