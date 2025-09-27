package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Ensure E_CALL_ARG_TYPE_MISMATCH points at the offending argument expression.
func TestCalls_ArgTypeMismatch_PositionPointsAtOffendingArg(t *testing.T) {
    // Callee expects (string,int); call passes (string,string); second arg should be flagged.
    src := "package app\nfunc Callee(a string, b int) {}\nfunc F(){ Callee(\"x\", \"y\") }\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeCalls(af)
    if len(ds) == 0 { t.Fatalf("expected diagnostics, got none") }
    var dpos *int
    for _, d := range ds {
        if d.Code == "E_CALL_ARG_TYPE_MISMATCH" && d.Pos != nil {
            col := d.Pos.Column
            dpos = &col
            break
        }
    }
    if dpos == nil { t.Fatalf("no E_CALL_ARG_TYPE_MISMATCH with position: %+v", ds) }
    // Expected column for second argument '"y"' is 24 on line 3 given our source layout.
    // func F(){ Callee("x", "y") }
    if *dpos != 24 { t.Fatalf("arg position column mismatch: got %d, want 24", *dpos) }
}
