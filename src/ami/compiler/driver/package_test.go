package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestPackage_StructFields(t *testing.T) {
    fs := &source.FileSet{}
    fs.AddFile("a.ami", "pipeline P{}")
    p := Package{Name: "main", Files: fs}
    if p.Name != "main" || p.Files == nil || p.Files.FileByName("a.ami") == nil {
        t.Fatalf("unexpected package: %+v", p)
    }
}

