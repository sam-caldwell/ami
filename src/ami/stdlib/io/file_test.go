package io

import (
    stdio "io"
    "os"
    "path/filepath"
    "testing"
)

func TestFHO_Create_Write_Read_Seek_Length_Pos_Flush_Close(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "sample.txt")

    h, err := Create(path)
    if err != nil { t.Fatalf("Create failed: %v", err) }

    data := []byte("hello world")
    n, err := h.Write(data)
    if err != nil { t.Fatalf("Write failed: %v", err) }
    if n != len(data) { t.Fatalf("Write bytes=%d want=%d", n, len(data)) }

    if err := h.Flush(); err != nil { t.Fatalf("Flush failed: %v", err) }

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
    if err != nil && err != stdio.EOF { t.Fatalf("Read failed: %v", err) }
    if rn != len(data) || string(buf) != string(data) {
        t.Fatalf("Read got=%d data=%q want=%q", rn, string(buf), string(data))
    }

    // Close and ensure operations error with ErrClosed
    if err := h.Close(); err != nil { t.Fatalf("Close failed: %v", err) }

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

func TestFHO_Open_OpenFile_Truncate(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "truncate.txt")

    // Prepare via os for flexibility
    f, err := os.Create(path)
    if err != nil { t.Fatalf("os.Create failed: %v", err) }
    if _, err := f.Write([]byte("ABCDEFGHIJ")); err != nil { t.Fatalf("seed write: %v", err) }
    f.Close()

    // Open read-only
    ro, err := Open(path)
    if err != nil { t.Fatalf("Open failed: %v", err) }
    b := make([]byte, 3)
    if n, err := ro.Read(b); err != nil && err != stdio.EOF { t.Fatalf("read ro: %v", err) } else if n != 3 {
        t.Fatalf("read bytes=%d want=3", n)
    }
    ro.Close()

    // OpenFile read-write and truncate
    rw, err := OpenFile(path, os.O_RDWR, 0o644)
    if err != nil { t.Fatalf("OpenFile failed: %v", err) }
    if err := rw.Truncate(4); err != nil { t.Fatalf("Truncate failed: %v", err) }
    if ln, err := rw.Length(); err != nil || ln != 4 {
        t.Fatalf("Length after truncate got=%d err=%v want=4", ln, err)
    }
    rw.Close()
}

