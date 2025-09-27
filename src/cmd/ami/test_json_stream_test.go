package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestJSONStream_RuntimeEventsAndSummary(t *testing.T) {
    dir := t.TempDir()
    // Create a simple runtime pass case
    if err := os.WriteFile(filepath.Join(dir, "r_test.ami"), []byte(`#pragma test:case ok
#pragma test:runtime input={"k":1} output={"k":1}
`), 0o644); err != nil { t.Fatal(err) }

    var out bytes.Buffer
    setTestOptions(TestOptions{Parallel: 1})
    if err := runTest(&out, dir, true, false, 0); err != nil {
        t.Fatalf("runTest error: %v\n%s", err, out.String())
    }
    // Scan NDJSON for test.v1 events and final summary object with runtime fields
    foundRunStart := false
    foundTestEnd := false
    foundRunEnd := false
    var last map[string]any
    sc := bufio.NewScanner(bytes.NewReader(out.Bytes()))
    for sc.Scan() {
        line := sc.Bytes()
        var m map[string]any
        if json.Unmarshal(line, &m) == nil {
            if m["schema"] == "test.v1" {
                switch m["type"] {
                case "run_start": foundRunStart = true
                case "test_end": foundTestEnd = true
                case "run_end": foundRunEnd = true
                }
            }
            last = m
        }
    }
    if !foundRunStart || !foundTestEnd || !foundRunEnd {
        t.Fatalf("missing runtime events: start=%v test_end=%v end=%v\n%s", foundRunStart, foundTestEnd, foundRunEnd, out.String())
    }
    // Verify final summary includes runtime fields
    if _, ok := last["runtime_tests"]; !ok { t.Fatalf("summary missing runtime_tests: %+v", last) }
    if _, ok := last["runtime_failures"]; !ok { t.Fatalf("summary missing runtime_failures: %+v", last) }
    if _, ok := last["runtime_skipped"]; !ok { t.Fatalf("summary missing runtime_skipped: %+v", last) }
}

