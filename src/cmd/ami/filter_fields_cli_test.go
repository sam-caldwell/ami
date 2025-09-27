package main

import (
    "bytes"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Verifies that --allow-field and --deny-field filter fields in debug logs when --verbose is on.
func TestRoot_AllowAndDenyFields_FilterDebugLog(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_filter_fields")
    if err := os.MkdirAll(dir, 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    cwd, _ := os.Getwd()
    defer os.Chdir(cwd)
    if err := os.Chdir(dir); err != nil { t.Fatalf("chdir: %v", err) }

    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    // allow only the "target" key; deny json and force explicitly
    c.SetArgs([]string{"--verbose", "--allow-field", "target", "--deny-field", "json", "--deny-field", "force", "init", "--json", "--force"})
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v", err) }
    closeRootLogger()

    data, err := os.ReadFile(filepath.Join("build", "debug", "activity.log"))
    if err != nil { t.Fatalf("read activity.log: %v", err) }
    s := string(data)
    // start line fields (dir,json,force) should be filtered away (allow only target)
    if strings.Contains(s, `"dir":`) {
        t.Fatalf("unexpected dir field present after allow filter: %s", s)
    }
    if strings.Contains(s, `"json":`) || strings.Contains(s, `"force":`) {
        t.Fatalf("unexpected json/force fields present after deny filter: %s", s)
    }
    // target_ready should still include target
    if !strings.Contains(s, `"message":"init.target_ready"`) || !strings.Contains(s, `"target":"./build"`) {
        t.Fatalf("expected target field retained for init.target_ready: %s", s)
    }
    // pkgroot_ready should have root dropped due to allow-only target
    if strings.Contains(s, `"root":`) {
        t.Fatalf("unexpected root field present with allow filter: %s", s)
    }
}

