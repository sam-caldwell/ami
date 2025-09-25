package parser

import (
	tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
	"testing"
)

func TestParser_Package_WithVersion(t *testing.T) {
	src := "package main:0.1.2\n"
	p := New(src)
	if p.cur.Kind != tok.KW_PACKAGE {
		t.Fatalf("cur kind=%v", p.cur.Kind)
	}
	p.next()
	if p.cur.Kind != tok.IDENT || p.cur.Lexeme != "main" {
		t.Fatalf("after package, got=%v %q", p.cur.Kind, p.cur.Lexeme)
	}
	p.next()
	if p.cur.Kind != tok.COLON {
		t.Fatalf("expected colon, got %v (%q)", p.cur.Kind, p.cur.Lexeme)
	}
	// reset parser to start position after our peeking
	p = New(src)
	f := p.ParseFile()
	if f.Package != "main" {
		t.Fatalf("package=%q", f.Package)
	}
	if f.Version != "0.1.2" {
		t.Fatalf("version=%q", f.Version)
	}
	// no parse errors expected for valid semver
	if len(p.Errors()) != 0 {
		t.Fatalf("unexpected errors: %+v", p.Errors())
	}
}

func TestParser_Package_WithVersionAndVPrefix(t *testing.T) {
	src := "package util:v1.2.3\n"
	p := New(src)
	f := p.ParseFile()
	if f.Package != "util" || f.Version != "v1.2.3" {
		t.Fatalf("got %q:%q", f.Package, f.Version)
	}
}

func TestParser_Package_BadVersion_ReportsDiagnostic(t *testing.T) {
	src := "package util:banana\n"
	p := New(src)
	_ = p.ParseFile()
	if len(p.Errors()) == 0 {
		t.Fatalf("expected version error")
	}
}
