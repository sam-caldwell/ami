package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppendLineIfMissing_AddsAndDedupe(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_init", "gitignore_unit")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	gi := filepath.Join(dir, ".gitignore")

	appendLineIfMissing(gi, "./build\n")
	b, err := os.ReadFile(gi)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(b) != "./build\n" {
		t.Fatalf("unexpected content: %q", string(b))
	}
	appendLineIfMissing(gi, "./build\n")
	b, _ = os.ReadFile(gi)
	if string(b) != "./build\n" {
		t.Fatalf("expected no duplicate lines, got: %q", string(b))
	}
}

func TestAppendLineIfMissing_ParentMissing_NoCreate(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_init", "gitignore_parent_missing")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	gi := filepath.Join(dir, "missing_parent", ".gitignore")
	appendLineIfMissing(gi, "./build\n")
	if _, err := os.Stat(gi); !os.IsNotExist(err) {
		t.Fatalf("expected no file created when parent missing; err=%v", err)
	}
}
