package astjson

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestToSchemaAST_PropagatesPackageVersion(t *testing.T) {
    src := "package util:1.2.3\n"
    p := parser.New(src)
    f := p.ParseFile()
    sch := ToSchemaAST(f, "u.ami")
    if sch.Version != "1.2.3" {
        t.Fatalf("schema version=%q", sch.Version)
    }
}

