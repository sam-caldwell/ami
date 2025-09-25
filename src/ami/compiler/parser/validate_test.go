package parser

import "testing"

func TestValidatePackageIdent(t *testing.T) {
	good := []string{"main", "pkg1", "_internal", "AlphaNum_123"}
	bad := []string{"1pkg", "has-dash", "", "a b", "."}
	for _, s := range good {
		if !ValidatePackageIdent(s) {
			t.Fatalf("expected valid: %q", s)
		}
	}
	for _, s := range bad {
		if ValidatePackageIdent(s) {
			t.Fatalf("expected invalid: %q", s)
		}
	}
}

func TestValidateImportPath(t *testing.T) {
	good := []string{
		"github.com/org/repo",
		"git.example.com/a.b_c-d/repo.v2",
		"./local/pkg",
	}
	bad := []string{
		"", "/abs/path", "./", "../up", "a//b", "a/./b", "a/../b", "a/b/../..", "white space/x", "a/b?c",
	}
	for _, s := range good {
		if !ValidateImportPath(s) {
			t.Fatalf("expected valid import: %q", s)
		}
	}
	for _, s := range bad {
		if ValidateImportPath(s) {
			t.Fatalf("expected invalid import: %q", s)
		}
	}
}
