package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// TestParser_Broader_Coverage exercises imports with constraints, literals, and pipelines with attrs.
func TestParser_Broader_Coverage(t *testing.T) {
    src := "package app\n" +
        "import foo >= v1.2.3\n" +
        "import (\n\t\"bar\" >= v0.1.0\n)\n" +
        "func F(){\n" +
        " var a slice<int> = slice<int>{1,2};\n" +
        " var b set<string> = set<string>{\"x\",\"y\"};\n" +
        " var c map<string,int> = map<string,int>{\"k1\":1, \"k2\":2};\n" +
        " a = a; b = b; c = c;\n" +
        "}\n" +
        "pipeline P(){ Alpha().Transform(\"W\").Collect edge.MultiPath(merge.Sort(\"ts\"), merge.Stable()); egress }\n" +
        "error { Foo() }\n"
    f := (&source.FileSet{}).AddFile("cov.ami", src)
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("ParseFile: %v", err) }
}

func TestParser_Pipeline_EdgeArrow(t *testing.T) {
    src := "package app\npipeline P(){ Alpha -> Beta; egress }\n"
    f := (&source.FileSet{}).AddFile("edge.ami", src)
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("ParseFile: %v", err) }
}

func TestParser_ResultList_ErrorBranch(t *testing.T) {
    src := "package app\nfunc X() (int, ) { return }\n"
    f := (&source.FileSet{}).AddFile("res.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}
