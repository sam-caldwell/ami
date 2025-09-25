package filepath

import "testing"

func TestFilepath_NormalizeAndClean(t *testing.T) {
	got := Clean(`a\\b/./c/../d`)
	if got != "a/b/d" {
		t.Fatalf("Clean got %q", got)
	}
}

func TestFilepath_Join_Base_Dir_Ext(t *testing.T) {
	p := Join("a", `b\\c`, "file.txt")
	if p != "a/b/c/file.txt" {
		t.Fatalf("Join got %q", p)
	}
	if Base(p) != "file.txt" {
		t.Fatalf("Base got %q", Base(p))
	}
	if Dir(p) != "a/b/c" {
		t.Fatalf("Dir got %q", Dir(p))
	}
	if Ext(p) != ".txt" {
		t.Fatalf("Ext got %q", Ext(p))
	}
}
