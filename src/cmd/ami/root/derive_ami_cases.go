package root

import (
    "path/filepath"
    "strconv"
    "strings"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// deriveAmiCases builds test cases from top-level #pragma directives in the file.
func deriveAmiCases(file, pkg string, f *astpkg.File) []amiCase {
    var cases []amiCase
    cur := amiCase{}
    flush := func() {
        if cur.name == "" {
            return
        }
        cases = append(cases, cur)
        cur = amiCase{}
    }
    for _, d := range f.Directives {
        switch d.Name {
        case "test:case":
            if cur.name != "" { flush() }
            name := strings.TrimSpace(d.Payload)
            if strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"") {
                name = strings.Trim(name, "\"")
            }
            cur = amiCase{name: name, file: file, pkg: pkg}
        case "test:expect_no_errors":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            cur.expects = append(cur.expects, amiExpect{kind: "no_errors"})
        case "test:expect_error":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            cur.expects = append(cur.expects, parseExpectWithParams("error", d.Payload))
        case "test:expect_warn":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            cur.expects = append(cur.expects, parseExpectWithParams("warn", d.Payload))
        case "test:expect_errors":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            e := parseExpectWithParams("errors_count", d.Payload)
            if !e.countSet { e.countSet = true; e.count = 1 }
            cur.expects = append(cur.expects, e)
        case "test:expect_warnings":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            e := parseExpectWithParams("warnings_count", d.Payload)
            if !e.countSet { e.countSet = true; e.count = 1 }
            cur.expects = append(cur.expects, e)
        case "test:expect_no_warnings":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            cur.expects = append(cur.expects, amiExpect{kind: "no_warnings"})
        case "test:skip":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            cur.skipReason = strings.TrimSpace(d.Payload)
        }
    }
    if cur.name != "" { flush() }
    if len(cases) == 0 {
        // default case: no-errors expectation
        cases = append(cases, amiCase{name: filepath.Base(file), file: file, pkg: pkg, expects: []amiExpect{{kind: "no_errors"}}})
    }
    return cases
}

// parseExpectWithParams parses payload like: CODE [count=N] [msg~="substr"]
func parseExpectWithParams(kind, payload string) amiExpect {
    e := amiExpect{kind: kind}
    fields := strings.Fields(payload)
    // The first field may be a code for error/warn kinds
    if (kind == "error" || kind == "warn") && len(fields) > 0 && !strings.Contains(fields[0], "=") && !strings.Contains(fields[0], "~") {
        e.code = fields[0]
        fields = fields[1:]
    }
    for _, f := range fields {
        if strings.HasPrefix(f, "count=") {
            n := strings.TrimPrefix(f, "count=")
            if i, err := strconv.Atoi(n); err == nil && i >= 0 {
                e.countSet = true
                e.count = i
            }
        }
        if strings.HasPrefix(f, "msg~=") {
            m := strings.TrimPrefix(f, "msg~=")
            e.msgSubstr = strings.Trim(m, "\"'")
        }
    }
    return e
}

