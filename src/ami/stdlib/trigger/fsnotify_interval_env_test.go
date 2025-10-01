package trigger

import (
    stdos "os"
    "path/filepath"
    "testing"
    "time"
    amios "github.com/sam-caldwell/ami/src/ami/stdlib/os"
    amiio "github.com/sam-caldwell/ami/src/ami/stdlib/io"
)

func TestFsNotify_Interval_FromEnv(t *testing.T) {
    // Set very small interval to speed up detection
    prev := amios.GetEnv("AMI_FS_WATCH_INTERVAL_MS")
    _ = amios.SetEnv("AMI_FS_WATCH_INTERVAL_MS", "1")
    t.Cleanup(func(){ _ = amios.SetEnv("AMI_FS_WATCH_INTERVAL_MS", prev) })

    dir := t.TempDir()
    path := filepath.Join(dir, "f.txt")
    ch, stop, err := FsNotify(path, FsCreate)
    if err != nil { t.Fatalf("FsNotify: %v", err) }
    defer stop()
    // Create quickly and expect event within a short window
    fh, err := amiio.Create(path)
    if err != nil { t.Fatalf("create: %v", err) }
    defer fh.Close()
    select {
    case <-ch:
        // ok
    case <-time.After(200 * time.Millisecond):
        t.Fatalf("timeout waiting for create with fast watch interval")
    }
    _ = stdos.Remove(path)
}

