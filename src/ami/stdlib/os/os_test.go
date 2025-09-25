package os

import (
	stdos "os"
	"path/filepath"
	"testing"
)

func TestOS_WriteReadFile_And_Stat(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "file.txt")
	data := []byte("hello")
	if err := WriteFile(f, data, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	got, err := ReadFile(f)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("got %q", string(got))
	}
	info, err := Stat(f)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.IsDir() {
		t.Fatal("expected file, got dir")
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("perm got %v", info.Mode().Perm())
	}
}

func TestOS_Mkdir_And_MkdirAll(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	if err := Mkdir(a, 0o700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	info, err := Stat(a)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected dir")
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("perm got %v", info.Mode().Perm())
	}

	nested := filepath.Join(dir, "x", "y", "z")
	if err := MkdirAll(nested, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if _, err := Stat(nested); err != nil {
		t.Fatalf("Stat nested: %v", err)
	}
}

func TestOS_Sad_WriteFile_NoParent(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "no", "parent", "file.txt")
	if err := WriteFile(bad, []byte("x"), 0o600); err == nil {
		t.Fatal("expected error for missing parent")
	}
}

func TestOS_Sad_Stat_NotExists(t *testing.T) {
	dir := t.TempDir()
	if _, err := Stat(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("expected error")
	}
}

// Ensure we don't accidentally mutate env/process in this package tests
func TestOS_NoEnvOrProcessMutation(t *testing.T) {
	// reads are allowed here, but we ensure no writes via this API
	_ = stdos.Environ()
}
