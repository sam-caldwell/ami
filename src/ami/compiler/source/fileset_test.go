package source

import "testing"

func TestFileSet_AddFileFromSource_PositionMapping(t *testing.T) {
	fs := NewFileSet()
	f := fs.AddFileFromSource("a.ami", "first\nsecond\nthird\n")

	// start of file
	p := f.PositionFor(0)
	if p.Line != 1 || p.Column != 1 {
		t.Fatalf("pos=%v want line1 col1", p)
	}

	// start of second line (offset after first \n)
	off2 := f.LineStartOffset(2)
	if off2 <= 0 {
		t.Fatalf("expected start offset for line 2")
	}
	p2 := f.PositionFor(off2)
	if p2.Line != 2 || p2.Column != 1 {
		t.Fatalf("pos2=%v want line2 col1", p2)
	}

	// a middle character in "second" (e.g., 'c')
	p3 := f.PositionFor(off2 + 2)
	if p3.Line != 2 || p3.Column != 3 {
		t.Fatalf("pos3=%v want line2 col3", p3)
	}

	// clamp past end
	p4 := f.PositionFor(10_000)
	if p4.Line != 3 {
		t.Fatalf("pos4 line=%d want 3", p4.Line)
	}
}
