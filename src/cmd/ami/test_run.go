package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

// runTest executes `go test ./...` in dir. When verbose, writes:
// - build/test/test.log (full test output)
// - build/test/test.manifest (each test name in execution order: "<package> <test>")
func runTest(out io.Writer, dir string, jsonOut bool, verbose bool) error {
    buildDir := filepath.Join(dir, "build", "test")
    if verbose {
        _ = os.MkdirAll(buildDir, 0o755)
    }

    // Decide command: use -json always to capture structured events for manifest.
    // Keep stdout quiet unless verbose or error; logs go to files under build/test.
    cmd := exec.Command("go", "test", "-json", "./...")
    cmd.Dir = dir
    // Make runs reproducible; avoid prompts
    env := append(os.Environ(), "GOTRACEBACK=single")
    cmd.Env = env
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    // In JSON mode, stream go test events directly to out while tee'ing to buffer for artifacts
    if jsonOut {
        cmd.Stdout = io.MultiWriter(out, &stdout)
    } else {
        cmd.Stdout = &stdout
    }
    cmd.Stderr = &stderr
    err := cmd.Run()

    // If verbose, write test.log
    if verbose {
        _ = os.WriteFile(filepath.Join(buildDir, "test.log"), stdout.Bytes(), 0o644)
    }

    // Build counts by scanning the go test JSON stream we captured in stdout buffer.
    // We count unique (package,test) pairs for tests, unique packages that have tests,
    // and failures as number of events with Action=="fail" and a Test name.
    type counts struct{ packages, tests, failures int }
    computeCounts := func(b []byte) counts {
        var c counts
        if len(b) == 0 { return c }
        dec := json.NewDecoder(bytes.NewReader(b))
        seenTests := map[string]struct{}{}
        seenPkgs := map[string]struct{}{}
        for dec.More() {
            var ev goTestEvent
            if dec.Decode(&ev) != nil {
                // skip malformed lines
                _ = dec.Decode(&map[string]any{})
                continue
            }
            if ev.Test != "" && ev.Package != "" && ev.Action == "run" {
                key := ev.Package + "\x00" + ev.Test
                if _, ok := seenTests[key]; !ok {
                    seenTests[key] = struct{}{}
                    if _, had := seenPkgs[ev.Package]; !had { seenPkgs[ev.Package] = struct{}{} }
                }
            }
            if ev.Test != "" && ev.Action == "fail" {
                c.failures++
            }
        }
        c.tests = len(seenTests)
        c.packages = len(seenPkgs)
        return c
    }
    c := computeCounts(stdout.Bytes())

    // Build manifest from JSON stream when verbose
    if verbose {
        mf, _ := os.OpenFile(filepath.Join(buildDir, "test.manifest"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
        if mf != nil {
            defer mf.Close()
            dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
            bw := bufio.NewWriter(mf)
            for dec.More() {
                var ev goTestEvent
                if dec.Decode(&ev) == nil {
                    if ev.Action == "run" && ev.Test != "" && ev.Package != "" {
                        _, _ = bw.WriteString(ev.Package + " " + ev.Test + "\n")
                    }
                } else {
                    // Skip malformed lines
                    _ = dec.Decode(&map[string]any{})
                }
            }
            _ = bw.Flush()
        }
    }

    // Emit a brief human summary to stdout when not JSON
    if !jsonOut {
        if err == nil {
            fmt.Fprintln(out, "test: OK")
        }
    } else {
        // Emit final JSON summary record as a newline-delimited object with counts
        _ = json.NewEncoder(out).Encode(map[string]any{
            "ok":       err == nil,
            "packages": c.packages,
            "tests":    c.tests,
            "failures": c.failures,
        })
    }

    if err != nil {
        // Write stderr summary to process stderr per SPEC
        msg := strings.TrimSpace(stderr.String())
        if msg != "" { fmt.Fprintln(os.Stderr, msg) }
        return err
    }
    return nil
}
