package astjson

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestToSchemaAST_IncludesStructuralNodes(t *testing.T) {
	const MainAMIFile = "main.ami"
	src := `package main:0.0.1
import alias "x/y"
pipeline P { Ingress(cfg).Transform(f)->Egress(cfg) }
func f() {}`
	p := parser.New(src)
	f := p.ParseFile()
	sch := ToSchemaAST(f, MainAMIFile)
	if sch.Schema != token.LexAstV1 || sch.Root.Kind != token.LexFile {
		t.Fatalf("bad schema root: %+v", sch)
	}
	kinds := make(map[string]int)
	for _, c := range sch.Root.Children {
		kinds[c.Kind]++
	}
	if kinds[token.DeclPackage] != 1 {
		t.Fatalf("%s missing: %v", token.DeclPackage, kinds)
	}
	if kinds[token.DeclImport] < 1 {
		t.Fatalf("%s missing: %v", token.DeclImport, kinds)
	}
	if kinds[token.DeclPipeline] != 1 {
		t.Fatalf("%s missing: %v", token.DeclPipeline, kinds)
	}
	if kinds[token.DeclFunc] != 1 {
		t.Fatalf("%s missing: %v", token.DeclFunc, kinds)
	}
}
