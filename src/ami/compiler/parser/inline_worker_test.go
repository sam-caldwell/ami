package parser

import (
    "testing"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func findPipelineDecl(f *astpkg.File) (astpkg.PipelineDecl, bool) {
    for _, d := range f.Decls {
        if p, ok := d.(astpkg.PipelineDecl); ok { return p, true }
    }
    return astpkg.PipelineDecl{}, false
}

func TestParser_InlineWorker_FunctionLiteral_Attached(t *testing.T) {
    src := `package p
pipeline P {
  Transform(worker=func(ev Event<string>) Event<string> { return ev })
}`
    p := New(src)
    f := p.ParseFile()
    pipe, ok := findPipelineDecl(f)
    if !ok || len(pipe.Steps) == 0 {
        t.Fatalf("missing pipeline decl")
    }
    tr := pipe.Steps[0]
    if tr.Name == "" || tr.Attrs["worker"] == "" {
        t.Fatalf("missing worker attribute")
    }
    if tr.InlineWorker == nil {
        t.Fatalf("expected InlineWorker function literal parsed")
    }
}

