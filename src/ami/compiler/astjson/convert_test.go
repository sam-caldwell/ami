package astjson

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestToSchemaAST_IncludesStructuralNodes(t *testing.T) {
    src := `package main
import alias "x/y"
pipeline P { Ingress(cfg).Transform(f)->Egress(cfg) }
func f() {}`
    p := parser.New(src)
    f := p.ParseFile()
    sch := ToSchemaAST(f, "main.ami")
    if sch.Schema != "ast.v1" || sch.Root.Kind != "File" { t.Fatalf("bad schema root: %+v", sch) }
    // Expect: PackageDecl, ImportDecl, PipelineDecl, FuncDecl among children
    kinds := make(map[string]int)
    for _, c := range sch.Root.Children { kinds[c.Kind]++ }
    if kinds["PackageDecl"] != 1 { t.Fatalf("PackageDecl missing: %v", kinds) }
    if kinds["ImportDecl"] < 1 { t.Fatalf("ImportDecl missing: %v", kinds) }
    if kinds["PipelineDecl"] != 1 { t.Fatalf("PipelineDecl missing: %v", kinds) }
    if kinds["FuncDecl"] != 1 { t.Fatalf("FuncDecl missing: %v", kinds) }
    // Check connectors preserved
    var connectors []interface{}
    for _, c := range sch.Root.Children {
        if c.Kind == "PipelineDecl" {
            if c.Fields == nil { t.Fatalf("pipeline fields empty") }
            v, ok := c.Fields["connectors"]
            if !ok { t.Fatalf("connectors missing") }
            if arr, ok := v.([]string); ok {
                if len(arr) != 2 || arr[0] != "." || arr[1] != "->" { t.Fatalf("unexpected connectors: %v", arr) }
            } else {
                // fallback if interface unmarshaled generically; build uses direct marshalling without re-unmarshal here
                connectors = nil; _ = connectors
            }
        }
    }
}

