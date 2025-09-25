package driver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDriver_PragmaBackpressure_FlowsIntoASM(t *testing.T) {
	dir := t.TempDir()
	src := "#pragma backpressure drop\npackage p\nfunc main(){}\n"
	path := filepath.Join(dir, "main.ami")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Compile([]string{path}, Options{})
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	if len(res.ASM) != 1 {
		t.Fatalf("expected 1 asm unit; got %d", len(res.ASM))
	}
	asm := res.ASM[0].Text
	if !strings.Contains(asm, "; backpressure drop") {
		t.Fatalf("asm missing backpressure: %q", asm)
	}
}
