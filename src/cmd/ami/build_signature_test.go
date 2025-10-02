package main

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/sam-caldwell/ami/src/ami/workspace"
	"os"
	"path/filepath"
	"testing"
)

func testRunBuild_Signature_Verify_Sum_Happy(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_build", "sig_sum_happy")
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save ws: %v", err)
	}
	// minimal ami.sum
	sum := []byte(`{"schema":"ami.sum/v1","packages":{}}`)
	if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil {
		t.Fatalf("write sum: %v", err)
	}
	h := sha256.Sum256(sum)
	if err := os.WriteFile(filepath.Join(dir, "ami.sum.sig"), []byte(hex.EncodeToString(h[:])), 0o644); err != nil {
		t.Fatalf("write sig: %v", err)
	}
	if err := runBuild(os.Stdout, dir, true, false); err != nil {
		t.Fatalf("runBuild: %v", err)
	}
}

func testRunBuild_Signature_Verify_Sum_Sad(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_build", "sig_sum_sad")
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save ws: %v", err)
	}
	// minimal ami.sum
	sum := []byte(`{"schema":"ami.sum/v1","packages":{}}`)
	if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil {
		t.Fatalf("write sum: %v", err)
	}
	// wrong signature
	if err := os.WriteFile(filepath.Join(dir, "ami.sum.sig"), []byte("deadbeef"), 0o644); err != nil {
		t.Fatalf("write sig: %v", err)
	}
	if err := runBuild(os.Stdout, dir, true, false); err == nil {
		t.Fatalf("expected signature failure")
	}
}
