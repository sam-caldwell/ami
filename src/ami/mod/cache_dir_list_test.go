package mod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheDir_And_List_Sorted(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Path before creation
	path, err := CacheDirPath()
	if err != nil {
		t.Fatalf("CacheDirPath: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected cache dir to not exist yet")
	}

	// Ensure creation
	dir, err := CacheDir()
	if err != nil {
		t.Fatalf("CacheDir: %v", err)
	}
	if dir != path {
		t.Fatalf("dir/path mismatch: %s vs %s", dir, path)
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatalf("cache dir missing after create: %v", err)
	}

	// Seed entries out of order
	for _, d := range []string{"b@v1.0.0", "a@v1.0.0", "c@v1.0.0"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	got, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	want := []string{"a@v1.0.0", "b@v1.0.0", "c@v1.0.0"}
	if len(got) != len(want) {
		t.Fatalf("len: got %d want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order[%d]: got %q want %q (all=%v)", i, got[i], want[i], got)
		}
	}
}
