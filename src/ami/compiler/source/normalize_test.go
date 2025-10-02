package source

import "testing"

func testAddFile_NormalizesCRLF_And_AppendsFinalNewline(t *testing.T) {
	var fs FileSet
	f := fs.AddFile("t.ami", "line1\r\nline2")
	want := "line1\nline2\n"
	if f.Content != want {
		t.Fatalf("content normalized = %q; want %q", f.Content, want)
	}
}

func testAddFile_BareCR_Normalized(t *testing.T) {
	var fs FileSet
	f := fs.AddFile("t.ami", "a\rb\rc")
	want := "a\nb\nc\n"
	if f.Content != want {
		t.Fatalf("content normalized = %q; want %q", f.Content, want)
	}
}

func testAddFile_Empty_NoChange(t *testing.T) {
	var fs FileSet
	f := fs.AddFile("empty.ami", "")
	if f.Content != "" {
		t.Fatalf("empty content should remain empty, got %q", f.Content)
	}
}
