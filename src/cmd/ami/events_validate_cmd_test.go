package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
    "time"
)

func TestEventsValidate_HappyAndSad(t *testing.T) {
    dir := t.TempDir()
    good := filepath.Join(dir, "good.json")
    bad := filepath.Join(dir, "bad.json")
    // Good event
    e := ev.Event{ID: "1", Timestamp: time.Unix(1700000000, 0).UTC(), Attempt: 1, Trace: map[string]any{"k":"v"}, Payload: map[string]any{"p":1}}
    if b, err := json.Marshal(e); err != nil { t.Fatal(err) } else { os.WriteFile(good, b, 0o644) }
    // Bad event (missing id)
    if err := os.WriteFile(bad, []byte(`{"schema":"events.v1"}`), 0o644); err != nil { t.Fatal(err) }
    c := newRootCmd()
    c.SetArgs([]string{"events", "validate", "--file", good})
    if err := c.Execute(); err != nil { t.Fatalf("good validate: %v", err) }
    c = newRootCmd()
    c.SetArgs([]string{"events", "validate", "--file", bad})
    if err := c.Execute(); err == nil { t.Fatalf("expected error for bad event") }
}

