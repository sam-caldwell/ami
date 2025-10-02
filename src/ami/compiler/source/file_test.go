package source

import "testing"

func testFile_Pos_Happy(t *testing.T) {
	f := &File{Name: "t.ami", Content: "a\nb\nc"}
	// offsets: 0:a (1,1), 1:'\n', 2:'b'(2,1), 3:'\n', 4:'c'(3,1)
	p0 := f.Pos(0)
	if p0.Line != 1 || p0.Column != 1 {
		t.Fatalf("want (1,1), got (%d,%d)", p0.Line, p0.Column)
	}
	p2 := f.Pos(2)
	if p2.Line != 2 || p2.Column != 2 {
		t.Fatalf("want (2,2), got (%d,%d)", p2.Line, p2.Column)
	}
	p4 := f.Pos(4)
	if p4.Line != 3 || p4.Column != 2 {
		t.Fatalf("want (3,2), got (%d,%d)", p4.Line, p4.Column)
	}
}

func testFile_Pos_Bounds(t *testing.T) {
	var f *File
	if p := f.Pos(0); p.Line != 0 {
		t.Fatalf("nil file should return zero pos: %+v", p)
	}
	nf := &File{Name: "t.ami", Content: ""}
	if p := nf.Pos(-1); p.Line != 0 {
		t.Fatalf("negative offset should be zero: %+v", p)
	}
	if p := nf.Pos(1); p.Line != 0 {
		t.Fatalf("offset > len should be zero: %+v", p)
	}
}
