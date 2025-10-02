package os

import (
    goos "os"
    "path/filepath"
    "testing"
    "time"
)

func TestWatch_Create_Modify_Remove(t *testing.T) {
    dir := t.TempDir()
    target := filepath.Join(dir, "watched.txt")
    ch, stop := Watch(target, 10*time.Millisecond)
    defer stop()

    // Create
    if err := goos.WriteFile(target, []byte("a"), 0o644); err != nil { t.Fatalf("create: %v", err) }
    // Expect create event
    select {
    case e := <-ch:
        if e.Op != OpCreate { t.Fatalf("expected create, got %v", e.Op) }
    case <-time.After(1 * time.Second):
        t.Fatalf("timeout waiting for create event")
    }

    // Modify
    if err := goos.WriteFile(target, []byte("ab"), 0o644); err != nil { t.Fatalf("modify: %v", err) }
    select {
    case e := <-ch:
        if e.Op != OpModify { t.Fatalf("expected modify, got %v", e.Op) }
    case <-time.After(1 * time.Second):
        t.Fatalf("timeout waiting for modify event")
    }

    // Remove
    if err := goos.Remove(target); err != nil { t.Fatalf("remove: %v", err) }
    select {
    case e := <-ch:
        if e.Op != OpRemove { t.Fatalf("expected remove, got %v", e.Op) }
    case <-time.After(1 * time.Second):
        t.Fatalf("timeout waiting for remove event")
    }
}

