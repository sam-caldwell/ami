package parser

import "testing"

func TestParser_ErrorRecovery_CollectsMultipleErrors(t *testing.T) {
	src := `pipeline P; func ; import ;`
	p := New(src)
	_ = p.ParseFile()
	errs := p.Errors()
	if len(errs) < 3 {
		t.Fatalf("expected >=3 errors, got %d", len(errs))
	}
}
