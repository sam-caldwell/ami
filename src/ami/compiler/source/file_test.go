package source

import "testing"

func TestFile_LineStartOffset_Bounds(t *testing.T) {
    fs := NewFileSet()
    f := fs.AddFileFromSource("a.ami", "a\nb\nc\n")

    if off := f.LineStartOffset(0); off != -1 {
        t.Fatalf("line 0 should be -1, got %d", off)
    }
    if off := f.LineStartOffset(1); off != 0 {
        t.Fatalf("line 1 start should be 0, got %d", off)
    }
    if off := f.LineStartOffset(2); off <= 0 {
        t.Fatalf("line 2 start should be >0, got %d", off)
    }
    if off := f.LineStartOffset(99); off != -1 {
        t.Fatalf("out of range line should be -1, got %d", off)
    }
}

func TestFile_PositionFor_ClampAndColumns(t *testing.T) {
    fs := NewFileSet()
    f := fs.AddFileFromSource("b.ami", "hello\nworld\n")

    // negative clamps to start
    p := f.PositionFor(-5)
    if p.Line != 1 || p.Column != 1 {
        t.Fatalf("neg clamp: got %v want line1 col1", p)
    }
    // start of second line
    off2 := f.LineStartOffset(2)
    p2 := f.PositionFor(off2)
    if p2.Line != 2 || p2.Column != 1 {
        t.Fatalf("line2 start: got %v want line2 col1", p2)
    }
    // middle of second line
    p3 := f.PositionFor(off2 + 3)
    if p3.Line != 2 || p3.Column != 4 {
        t.Fatalf("line2 mid: got %v want line2 col4", p3)
    }
    // far past end clamps to last line
    p4 := f.PositionFor(1_000_000)
    if p4.Line != 2 {
        t.Fatalf("end clamp: got line %d want 2", p4.Line)
    }
}

