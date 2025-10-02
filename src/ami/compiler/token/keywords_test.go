package token

import "testing"

func testLookupKeyword_Hit(t *testing.T) {
	kind, ok := LookupKeyword("package")
	if !ok || kind != KwPackage {
		t.Fatalf("LookupKeyword(package) => (%v,%v); want (KwPackage,true)", kind, ok)
	}
}

func testLookupKeyword_Miss(t *testing.T) {
	kind, ok := LookupKeyword("notakeyword")
	if ok || kind != Ident {
		t.Fatalf("LookupKeyword(miss) => (%v,%v); want (Ident,false)", kind, ok)
	}
}
