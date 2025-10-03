package main

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestAttrsFromStep_CoversMergeTypeAndMultipath(t *testing.T) {
    // nil should yield nil
    if got := attrsFromStep(nil); got != nil {
        t.Fatalf("nil step should yield nil, got: %v", got)
    }

    // Build a step with various attributes to exercise all branches
    st := &ast.StepStmt{
        Attrs: []ast.Attr{
            // Type attribute (case-insensitive)
            {Name: "Type", Args: []ast.Arg{{Text: "Reducer"}}},
            // merge.Buffer with non-zero size and block policy ⇒ bounded + atLeastOnce
            {Name: "merge.Buffer", Args: []ast.Arg{{Text: "1"}, {Text: "block"}}},
            // merge.Shunt newest ⇒ shuntNewest (overrides delivery)
            {Name: "merge.Shunt", Args: []ast.Arg{{Text: "newest"}}},
            // MultiPath with multiple args ⇒ joined by '|'
            {Name: "edge.MultiPath", Args: []ast.Arg{{Text: "A"}, {Text: "B"}, {Text: ""}, {Text: "C"}}},
        },
    }
    got := attrsFromStep(st)
    if got == nil { t.Fatalf("expected non-nil attrs map") }
    if v, ok := got["bounded"]; !ok || v != true {
        t.Fatalf("expected bounded=true, got: %v (ok=%v)", v, ok)
    }
    if v, ok := got["delivery"]; !ok || v != "shuntNewest" {
        t.Fatalf("expected delivery=shuntNewest, got: %v (ok=%v)", v, ok)
    }
    if v, ok := got["type"]; !ok || v != "Reducer" {
        t.Fatalf("expected type=Reducer, got: %v (ok=%v)", v, ok)
    }
    if v, ok := got["multipath"]; !ok || v != "A|B|C" {
        t.Fatalf("expected multipath=A|B|C, got: %v (ok=%v)", v, ok)
    }
}
