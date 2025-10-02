package io

import (
	stdio "io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testFHO_Create_Write_Read_Seek_Length_Pos_Flush_Close(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")

	h, err := Create(path)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	data := []byte("hello world")
	n, err := h.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Write bytes=%d want=%d", n, len(data))
	}

	if err := h.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if got, err := h.Length(); err != nil || got != int64(len(data)) {
		t.Fatalf("Length got=%d err=%v want=%d", got, err, len(data))
	}

	if pos, err := h.Pos(); err != nil || pos != int64(len(data)) {
		t.Fatalf("Pos got=%d err=%v want=%d", pos, err, len(data))
	}

	if _, err := h.Seek(0, stdio.SeekStart); err != nil {
		t.Fatalf("Seek start failed: %v", err)
	}

	buf := make([]byte, len(data))
	rn, err := h.Read(buf)
	if err != nil && err != stdio.EOF {
		t.Fatalf("Read failed: %v", err)
	}
	if rn != len(data) || string(buf) != string(data) {
		t.Fatalf("Read got=%d data=%q want=%q", rn, string(buf), string(data))
	}
	// Reading again should yield 0, EOF (partial read semantics with non-nil error)
	more := make([]byte, 4)
	r2, err := h.Read(more)
	if err == nil || err != stdio.EOF || r2 != 0 {
		t.Fatalf("Expected EOF on subsequent read, got n=%d err=%v", r2, err)
	}

	// Close and ensure operations error with ErrClosed
	if err := h.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if _, err := h.Write([]byte("x")); err == nil || err != ErrClosed {
		t.Fatalf("Write after close err=%v want ErrClosed", err)
	}
	if _, err := h.Read(make([]byte, 1)); err == nil || err != ErrClosed {
		t.Fatalf("Read after close err=%v want ErrClosed", err)
	}
	if _, err := h.Seek(0, stdio.SeekCurrent); err == nil || err != ErrClosed {
		t.Fatalf("Seek after close err=%v want ErrClosed", err)
	}
	if err := h.Flush(); err == nil || err != ErrClosed {
		t.Fatalf("Flush after close err=%v want ErrClosed", err)
	}
}

func testFHO_Open_OpenFile_Truncate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "truncate.txt")

	// Prepare via os for flexibility
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("os.Create failed: %v", err)
	}
	if _, err := f.Write([]byte("ABCDEFGHIJ")); err != nil {
		t.Fatalf("seed write: %v", err)
	}
	f.Close()

	// Open read-only
	ro, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	b := make([]byte, 3)
	if n, err := ro.Read(b); err != nil && err != stdio.EOF {
		t.Fatalf("read ro: %v", err)
	} else if n != 3 {
		t.Fatalf("read bytes=%d want=3", n)
	}
	ro.Close()

	// OpenFile read-write and truncate
	rw, err := OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	if err := rw.Truncate(4); err != nil {
		t.Fatalf("Truncate failed: %v", err)
	}
	if ln, err := rw.Length(); err != nil || ln != 4 {
		t.Fatalf("Length after truncate got=%d err=%v want=4", ln, err)
	}
	rw.Close()
}

func testCreateTemp_CreateTempDir_Stat_Name(t *testing.T) {
	// Create temp file with suffix
	fho, err := CreateTemp("ami-tests", ".log")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	defer fho.Close()
	if n, err := fho.Write([]byte("abc")); err != nil || n != 3 {
		t.Fatalf("Write: n=%d err=%v", n, err)
	}
	if err := fho.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	name := fho.Name()
	if name == "" {
		t.Fatalf("Name() should not be empty")
	}
	if !strings.HasSuffix(name, ".log") {
		t.Fatalf("temp file suffix missing: %s", name)
	}

	info, err := Stat(name)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Size != 3 || info.IsDir {
		t.Fatalf("Stat unexpected: %+v", info)
	}

	// Temp dir
	dir, err := CreateTempDir()
	if err != nil {
		t.Fatalf("CreateTempDir failed: %v", err)
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatalf("temp dir invalid: %v isDir=%v err=%v", dir, fi.IsDir(), err)
	}
}

func testStdStreams_CloseOnlyMarksHandle(t *testing.T) {
	// Write a small message to stdout/stderr via FHO, then Close and ensure further writes fail
	if n, err := Stdout.Write([]byte("")); err != nil || n != 0 {
		t.Fatalf("Stdout write zero bytes err=%v n=%d", err, n)
	}
	if n, err := Stderr.Write([]byte("")); err != nil || n != 0 {
		t.Fatalf("Stderr write zero bytes err=%v n=%d", err, n)
	}

	if err := Stdout.Close(); err != nil {
		t.Fatalf("Stdout close: %v", err)
	}
	if err := Stderr.Close(); err != nil {
		t.Fatalf("Stderr close: %v", err)
	}

	if _, err := Stdout.Write([]byte("x")); err == nil || err != ErrClosed {
		t.Fatalf("Stdout expected ErrClosed, got %v", err)
	}
	if _, err := Stderr.Write([]byte("x")); err == nil || err != ErrClosed {
		t.Fatalf("Stderr expected ErrClosed, got %v", err)
	}
}

func testNameNilAndTempVariantsAndClosedTruncate(t *testing.T) {
	// Name on nil receiver returns empty
	var hnil *FHO
	if name := hnil.Name(); name != "" {
		t.Fatalf("nil Name() => %q, want empty", name)
	}

	// CreateTemp with only dir argument
	fho, err := CreateTemp("ami-tests2")
	if err != nil {
		t.Fatalf("CreateTemp(dir) failed: %v", err)
	}
	path := fho.Name()
	if !strings.Contains(path, string(os.PathSeparator)+"ami-tests2"+string(os.PathSeparator)) {
		t.Fatalf("CreateTemp dir not applied: %s", path)
	}
	fho.Close()

	// invalid args (>2)
	if _, err := CreateTemp("a", "b", "c"); err == nil {
		t.Fatalf("CreateTemp expected error on too many args")
	}

	// Truncate on closed handle returns ErrClosed
	fho2, err := CreateTemp()
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	if err := fho2.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := fho2.Truncate(0); err == nil || err != ErrClosed {
		t.Fatalf("Truncate after close expected ErrClosed, got %v", err)
	}
}
