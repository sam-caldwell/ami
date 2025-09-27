package main

import (
    "os"
    "path/filepath"
    "testing"
    "bytes"
)

func TestRuntimeCLI_BasicPassFailSkip(t *testing.T) {
    dir := t.TempDir()
    // create runtime test files
    mustWrite := func(rel, content string){
        p := filepath.Join(dir, rel)
        if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil { t.Fatal(err) }
        if err := os.WriteFile(p, []byte(content), 0o644); err != nil { t.Fatal(err) }
    }
    // pass: identity
    mustWrite("pass_test.ami", `
#pragma test:case pass
#pragma test:runtime input={"k":1} output={"k":1}
`)
    // fail: expect ok but inject error_code (no expect_error) -> should fail
    mustWrite("fail_test.ami", `
#pragma test:case fail
#pragma test:runtime input={"error_code":"E_FAIL"}
`)
    // skip: marked skipped
    mustWrite("skip_test.ami", `
#pragma test:case skip
#pragma test:skip because
#pragma test:runtime input={"k":2} output={"k":2}
`)

    var out bytes.Buffer
    setTestOptions(TestOptions{Parallel: 1})
    if err := runTest(&out, dir, false, false, 0); err == nil {
        // Should be non-nil because one runtime case fails
        t.Fatalf("expected failure due to runtime fail; out=\n%s", out.String())
    }
    s := out.String()
    if !bytes.Contains([]byte(s), []byte("test: runtime ok=1 fail=1 skip=1")) {
        t.Fatalf("summary missing or incorrect: %s", s)
    }
}

