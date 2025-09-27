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
    "sort"
    "strings"
    "strconv"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// runTest executes `go test ./...` in dir. When verbose, writes:
// - build/test/test.log (full test output)
// - build/test/test.manifest (each test name in execution order: "<package> <test>")
// It also evaluates AMI directive-based tests in any `*.ami` files under dir:
// - `#pragma test:case <name>` defines a case
// - `#pragma test:assert parse_ok|parse_fail` sets expected parse outcome (default parse_ok)
// Failing AMI directive tests cause a non-zero exit and are included in summaries.
func runTest(out io.Writer, dir string, jsonOut bool, verbose bool, pkgConcurrency int) error {
    buildDir := filepath.Join(dir, "build", "test")
    if verbose {
        _ = os.MkdirAll(buildDir, 0o755)
    }

    // Decide command: use -json always to capture structured events for manifest.
    // Keep stdout quiet unless verbose or error; logs go to files under build/test.
    // Build arguments for `go test`.
    args := []string{"test", "-json"}
    if pkgConcurrency > 0 { args = append(args, "-p", fmt.Sprintf("%d", pkgConcurrency)) }
    args = append(args, "./...")
    cmd := exec.Command("go", args...)
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
    type counts struct{
        packages, tests, failures int
        byPkg map[string]struct{ ok, fail int }
    }
    computeCounts := func(b []byte) counts {
        var c counts
        c.byPkg = map[string]struct{ ok, fail int }{}
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
            if ev.Test != "" && ev.Package != "" {
                entry := c.byPkg[ev.Package]
                switch ev.Action {
                case "pass": entry.ok++
                case "fail": entry.fail++; c.failures++
                }
                c.byPkg[ev.Package] = entry
            }
        }
        c.tests = len(seenTests)
        c.packages = len(seenPkgs)
        return c
    }
    c := computeCounts(stdout.Bytes())
    if err != nil && c.packages == 0 {
        // Tolerate empty Go package trees; AMI directive tests may still run.
        s := stdout.String() + "\n" + stderr.String()
        if strings.Contains(s, "no packages to test") || strings.Contains(s, "matched no packages") {
            err = nil
            stderr.Reset()
        }
    }

    // Run AMI directive tests by scanning *.ami files.
    type amiResult struct{
        name   string
        file   string
        ok     bool
        expect string
        count  int
        msgSub string
        gotErrs int
    }
    amiTests, amiFailures := 0, 0
    var amiManifest []string
    var amiEvents []amiResult
    evaluateAMIDirectives := func(root string) {
        // Walk tree shallowly by listing files recursively using filepath.WalkDir.
        _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
            if err != nil { return nil }
            if d.IsDir() { return nil }
            if !strings.HasSuffix(d.Name(), ".ami") { return nil }
            // Read file
            b, err := os.ReadFile(path)
            if err != nil { return nil }
            // Parse using compiler/parser
            var fs source.FileSet
            f := fs.AddFile(path, string(b))
            p := parser.New(f)
            file, errs := p.ParseFileCollect()
            // default expectation: parse_ok
            expectParseOK := true
            expCount := -1
            expMsg := ""
            var caseNames []string
            for _, pr := range file.Pragmas {
                if pr.Domain == "test" {
                    switch pr.Key {
                    case "case":
                        if pr.Value != "" { caseNames = append(caseNames, pr.Value) }
                    case "assert":
                        if len(pr.Args) > 0 {
                            if pr.Args[0] == "parse_fail" { expectParseOK = false }
                            if pr.Args[0] == "parse_ok" { expectParseOK = true }
                        }
                        if pr.Params != nil {
                            if v, ok := pr.Params["count"]; ok {
                                if n, e := strconv.Atoi(v); e == nil { expCount = n }
                            }
                            if v, ok := pr.Params["msg"]; ok { expMsg = v }
                            if v, ok := pr.Params["message"]; ok { expMsg = v }
                        }
                    }
                }
            }
            if len(caseNames) == 0 { return nil }
            // Determine result based on expectation vs error
            ok := (len(errs) == 0) == expectParseOK
            // Count assertion
            if expCount >= 0 && ok {
                if expectParseOK && expCount != 0 { ok = false }
                if !expectParseOK && expCount != len(errs) { ok = false }
            }
            // Message substring assertion when provided
            if expMsg != "" && ok {
                s := ""
                for _, e := range errs { s += e.Error() + "\n" }
                if !strings.Contains(s, expMsg) { ok = false }
            }
            // Record for each case
            rel := path
            if rp, rerr := filepath.Rel(root, path); rerr == nil { rel = rp }
            for _, name := range caseNames {
                amiTests++
                if !ok { amiFailures++ }
                amiManifest = append(amiManifest, fmt.Sprintf("ami:%s %s", rel, name))
                amiEvents = append(amiEvents, amiResult{name: name, file: rel, ok: ok, expect: map[bool]string{true:"parse_ok", false:"parse_fail"}[expectParseOK], count: expCount, msgSub: expMsg, gotErrs: len(errs)})
            }
            return nil
        })
    }
    evaluateAMIDirectives(dir)

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
            // Append AMI directive test entries, sorted for determinism.
            sort.Strings(amiManifest)
            for _, line := range amiManifest { _, _ = bw.WriteString(line + "\n") }
            _ = bw.Flush()
        }
    }

    // Emit a brief human summary to stdout when not JSON
    if !jsonOut {
        // Per-package summaries
        if len(c.byPkg) > 0 {
            // stable order
            pkgs := make([]string, 0, len(c.byPkg))
            for k := range c.byPkg { pkgs = append(pkgs, k) }
            sort.Strings(pkgs)
            for _, pkg := range pkgs {
                v := c.byPkg[pkg]
                fmt.Fprintf(out, "test: pkg %s ok=%d fail=%d\n", pkg, v.ok, v.fail)
            }
        }
        if amiTests > 0 {
            fmt.Fprintf(out, "test: ami ok=%d fail=%d\n", amiTests-amiFailures, amiFailures)
        }
        if err == nil && amiFailures == 0 {
            fmt.Fprintln(out, "test: OK")
        }
    } else {
        // Emit final JSON summary record as a newline-delimited object with counts
        // Stream AMI directive events before the final summary for visibility.
        for _, ev := range amiEvents {
            _ = json.NewEncoder(out).Encode(map[string]any{
                "schema":  "ami.test.v1",
                "file":    ev.file,
                "case":    ev.name,
                "ok":      ev.ok,
                "expect":  ev.expect,
                "count":   ev.count,
                "gotErrs": ev.gotErrs,
                "msg":     ev.msgSub,
            })
        }
        _ = json.NewEncoder(out).Encode(map[string]any{
            "ok":       err == nil && amiFailures == 0,
            "packages": c.packages,
            "tests":    c.tests,
            "failures": c.failures,
            "ami_tests":    amiTests,
            "ami_failures": amiFailures,
        })
    }

    if err != nil || amiFailures > 0 {
        // Write stderr summary to process stderr per SPEC
        msg := strings.TrimSpace(stderr.String())
        if msg != "" { fmt.Fprintln(os.Stderr, msg) }
        if err != nil { return err }
        return fmt.Errorf("ami directive tests failed")
    }
    return nil
}
