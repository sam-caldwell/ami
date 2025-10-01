package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParse_ErrorBlock_RichSteps(t *testing.T) {
    src := "package app\nerror {\n// lead comment\nAlpha(@, 1) edge.MultiPath(, merge.Stable()) ;\n} \n"
    f := (&source.FileSet{}).AddFile("richerr.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}

