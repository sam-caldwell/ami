package parser

import (
    "testing"
)

// Ensure parser rejects pointer type syntax with E_PTR_UNSUPPORTED_SYNTAX.
func TestParser_PointerTypeSyntax_Rejected(t *testing.T) {
    src := `package p
func f(p *int) { }
`
    p := New(src)
    _ = p.ParseFile()
    errs := p.Errors()
    if len(errs) == 0 { t.Fatalf("expected E_PTR_UNSUPPORTED_SYNTAX; got none") }
    found := false
    for _, e := range errs { if e.Code == "E_PTR_UNSUPPORTED_SYNTAX" { found = true; break } }
    if !found { t.Fatalf("expected E_PTR_UNSUPPORTED_SYNTAX; errs=%+v", errs) }
}

// Ensure parser rejects address-of syntax with E_PTR_UNSUPPORTED_SYNTAX.
func TestParser_AddressOfSyntax_Rejected(t *testing.T) {
    src := `package p
func f() {
  mutate(p = &x)
}
`
    p := New(src)
    _ = p.ParseFile()
    errs := p.Errors()
    if len(errs) == 0 { t.Fatalf("expected E_PTR_UNSUPPORTED_SYNTAX; got none") }
    found := false
    for _, e := range errs { if e.Code == "E_PTR_UNSUPPORTED_SYNTAX" { found = true; break } }
    if !found { t.Fatalf("expected E_PTR_UNSUPPORTED_SYNTAX; errs=%+v", errs) }
}

