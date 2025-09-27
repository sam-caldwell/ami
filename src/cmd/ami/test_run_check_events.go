package main

import (
    "fmt"
    "io"
    "time"

    "github.com/sam-caldwell/ami/src/schemas/events"
)

// runCheckEvents validates a sample events.Event to exercise the schema path.
// It is intentionally a no-op for users and only runs when --check-events is provided.
func runCheckEvents(out io.Writer) error {
    e := events.Event{ID: "probe", Timestamp: time.Now().UTC(), Attempt: 1}
    if err := events.Validate(e); err != nil {
        // Surface minimal info; keep silent by default otherwise.
        _, _ = fmt.Fprintln(out, "events: validation failed:", err)
        return err
    }
    // Print nothing on success to avoid clutter.
    return nil
}

