package main

import (
    "bufio"
    "encoding/json"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

// parseRuntimeCases scans for `*_test.ami` files and collects runtime cases based on pragmas.
// Simplified rule: a file must include at least one `#pragma test:case <name>` and a single
// `#pragma test:runtime ...` block that applies to all cases in the file. `test:skip` applies to all cases.
func parseRuntimeCases(root string) ([]runtimeCase, error) {
    var cases []runtimeCase
    err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
        if err != nil || d.IsDir() { return nil }
        if filepath.Ext(p) != ".ami" || !strings.HasSuffix(p, "_test.ami") { return nil }
        f, err := os.Open(p); if err != nil { return nil }
        defer f.Close()
        var names []string
        spec := runtimeSpec{}
        scan := bufio.NewScanner(f)
        for scan.Scan() {
            line := strings.TrimSpace(scan.Text())
            if !strings.HasPrefix(line, "#pragma test:") { continue }
            body := strings.TrimSpace(strings.TrimPrefix(line, "#pragma test:"))
            if strings.HasPrefix(body, "case ") {
                name := strings.TrimSpace(strings.TrimPrefix(body, "case "))
                if name != "" { names = append(names, name) }
            } else if strings.HasPrefix(body, "skip") {
                reason := strings.TrimSpace(strings.TrimPrefix(body, "skip"))
                if reason == "" { reason = "skipped" }
                spec.SkipReason = strings.TrimSpace(reason)
            } else if strings.HasPrefix(body, "fixture ") {
                rest := strings.TrimSpace(strings.TrimPrefix(body, "fixture "))
                kv := parseKV(rest)
                if pth := kv["path"]; pth != "" {
                    mode := kv["mode"]
                    if mode != "rw" { mode = "ro" }
                    spec.Fixtures = append(spec.Fixtures, fixtureSpec{Path: pth, Mode: mode})
                }
            } else if strings.HasPrefix(body, "runtime ") {
                rest := strings.TrimSpace(strings.TrimPrefix(body, "runtime "))
                kv := parseKV(rest)
                if v := kv["input"]; v != "" { spec.InputJSON = v }
                if v := kv["output"]; v != "" { spec.ExpectJSON = v }
                if v := kv["expect_error"]; v != "" { spec.ExpectError = v }
                if v := kv["timeout"]; v != "" { if n, e := strconv.Atoi(v); e == nil { spec.TimeoutMs = n } }
            }
        }
        if len(names) == 0 { return nil }
        // Validate JSON snippets early
        if spec.InputJSON != "" { var tmp any; _ = json.Unmarshal([]byte(spec.InputJSON), &tmp) }
        if rel, e := filepath.Rel(root, p); e == nil { p = rel }
        for _, n := range names { cases = append(cases, runtimeCase{File: p, Name: n, Spec: spec}) }
        return nil
    })
    return cases, err
}

// parseKV parses a simple space-delimited list of key=value tokens. Values may be wrapped in single or double quotes.
func parseKV(s string) map[string]string {
    out := map[string]string{}
    parts := strings.Fields(s)
    for _, p := range parts {
        if i := strings.IndexByte(p, '='); i > 0 {
            k := p[:i]
            v := p[i+1:]
            v = strings.Trim(v, `"'`)
            out[k] = v
        }
    }
    return out
}
