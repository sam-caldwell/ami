package parser

import "testing"

func TestParser_CollectsPragmas(t *testing.T) {
	src := `#pragma concurrency 4
#pragma capabilities net,fs
#pragma trust sandboxed
#pragma backpressure drop
package p`
	p := New(src)
	f := p.ParseFile()
	if len(f.Directives) != 4 {
		t.Fatalf("want 4 directives; got %d", len(f.Directives))
	}
	if f.Directives[0].Name != "concurrency" {
		t.Fatalf("first directive name=%q", f.Directives[0].Name)
	}
	if f.Directives[3].Name != "backpressure" || f.Directives[3].Payload != "drop" {
		t.Fatalf("backpressure payload mismatch: %+v", f.Directives[3])
	}
}

func TestParser_PackageAndImportValidation_Diagnostics(t *testing.T) {
	src := `package 9bad
import "../bad"
`
	p := New(src)
	_ = p.ParseFile()
	errs := p.Errors()
	if len(errs) < 2 {
		t.Fatalf("expected at least 2 diagnostics; got %d", len(errs))
	}
	var seenPkg, seenImp bool
	for _, e := range errs {
		if e.Code == "E_BAD_PACKAGE" {
			seenPkg = true
		}
		if e.Code == "E_BAD_IMPORT" {
			seenImp = true
		}
	}
	if !seenPkg || !seenImp {
		t.Fatalf("missing diag codes; got %+v", errs)
	}
}
