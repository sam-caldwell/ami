package diag

import (
	"testing"
)

func TestDiagnostic_ToSchema_Validate(t *testing.T) {
	d := Diagnostic{Level: Error, Code: "E_TEST", Message: "msg", Package: "pkg", File: "f.ami"}
	s := d.ToSchema()
	if err := s.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if s.Level != "error" || s.Code != "E_TEST" || s.File != "f.ami" {
		t.Fatalf("unexpected fields: %+v", s)
	}
}
