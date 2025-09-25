package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

// Verify assignment uses generic unification so Event<T> := Event<string> passes.
func TestAssignment_GenericUnify_OK(t *testing.T) {
    src := `package p
func f(x Event<T>, y Event<string>) { *x = y }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics {
        if d.Code == "E_ASSIGN_TYPE_MISMATCH" {
            t.Fatalf("unexpected assign mismatch: %v", d)
        }
    }
}

