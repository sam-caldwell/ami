package llvm

import "testing"

func testToolError_ErrorString(t *testing.T) {
	e := ToolError{Tool: "clang", Stderr: "boom"}
	if got := e.Error(); got != "clang failed" {
		t.Fatalf("unexpected error string: %q", got)
	}
}

func testVersion_PathEmptyOrClang(t *testing.T) {
	if _, err := Version(""); err == nil {
		t.Fatalf("expected error on empty path")
	}
	// If clang is available, ensure happy path returns a non-empty string; otherwise skip.
	if clang, err := FindClang(); err == nil {
		if ver, err := Version(clang); err != nil || ver == "" {
			t.Fatalf("expected clang version, got ver=%q err=%v", ver, err)
		}
	}
}
