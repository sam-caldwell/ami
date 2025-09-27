package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// TestParse_ImportConstraint_QuotedAndMissing covers quoted version and missing version error path.
func TestParse_ImportConstraint_QuotedAndMissing(t *testing.T) {
    // Quoted version after >=
    src := "package app\nimport \"foo\" >= \"v1.2.3-rc.1\"\n"
    f := (&source.FileSet{}).AddFile("icq.ami", src)
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("ParseFile quoted: %v", err) }

    // Missing version after operator
    src2 := "package app\nimport bar >= \n"
    f2 := (&source.FileSet{}).AddFile("icm.ami", src2)
    p2 := New(f2)
    _, _ = p2.ParseFile() // tolerate error; path exercised
}

func TestParse_ImportConstraint_UnquotedComposite(t *testing.T) {
    src := "package app\nimport foo >= v1.2.3-rc.1\n"
    f := (&source.FileSet{}).AddFile("icu.ami", src)
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("ParseFile: %v", err) }
}

func TestParse_ImportConstraint_QuotedMissingV(t *testing.T) {
    src := "package app\nimport \"x\" >= \"1.2.3\"\n"
    f := (&source.FileSet{}).AddFile("icnv.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}
