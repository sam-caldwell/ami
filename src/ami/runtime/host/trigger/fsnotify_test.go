package trigger

import (
    stdos "os"
    "path/filepath"
    "testing"
    "time"
    amiio "github.com/sam-caldwell/ami/src/ami/runtime/host/io"
)

func TestFsNotify_Create_Modify_Remove(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "f.txt")

    // Create
    ch1, stop1, err := FsNotify(path, FsCreate)
    if err != nil { t.Fatalf("FsNotify create: %v", err) }
    defer stop1()
    // trigger create
    fh, err := amiio.Create(path)
    if err != nil { t.Fatalf("create file: %v", err) }
    defer fh.Close()
    select {
    case e := <-ch1:
        if e.Value.Op != FsCreate { t.Fatalf("expected FsCreate, got %v", e.Value.Op) }
        if e.Value.Handle == nil { t.Fatalf("expected handle on create") }
    case <-time.After(1 * time.Second):
        t.Fatalf("timeout waiting for create event")
    }

    // Modify
    ch2, stop2, err := FsNotify(path, FsModify)
    if err != nil { t.Fatalf("FsNotify modify: %v", err) }
    defer stop2()
    if _, err := fh.Write([]byte("x")); err != nil { t.Fatalf("write: %v", err) }
    if err := fh.Flush(); err != nil { t.Fatalf("flush: %v", err) }
    select {
    case e := <-ch2:
        if e.Value.Op != FsModify { t.Fatalf("expected FsModify, got %v", e.Value.Op) }
    case <-time.After(1 * time.Second):
        t.Fatalf("timeout waiting for modify event")
    }

    // Remove
    ch3, stop3, err := FsNotify(path, FsRemove)
    if err != nil { t.Fatalf("FsNotify remove: %v", err) }
    defer stop3()
    if err := stdos.Remove(path); err != nil { t.Fatalf("remove: %v", err) }
    select {
    case e := <-ch3:
        if e.Value.Op != FsRemove { t.Fatalf("expected FsRemove, got %v", e.Value.Op) }
        if e.Value.Handle != nil { t.Fatalf("expected nil handle on remove") }
    case <-time.After(1 * time.Second):
        t.Fatalf("timeout waiting for remove event")
    }
}

