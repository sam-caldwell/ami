package root

import (
    "path/filepath"
    "strconv"
    "strings"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// ---- AMI runtime tests (Phase 2: deterministic harness integration) ----

// deriveAmiRuntimeCases builds runtime cases from `#pragma test:runtime ...` directives.
// file basename is used. Each `test:runtime` directive flushes a case.
func deriveAmiRuntimeCases(file, pkg string, f *astpkg.File) []amiRuntimeCase {
    var cases []amiRuntimeCase
    var curName string
    var curFixtures []amiFixture
    for _, d := range f.Directives {
        switch d.Name {
        case "test:case":
            curName = strings.TrimSpace(d.Payload)
            if strings.HasPrefix(curName, "\"") && strings.HasSuffix(curName, "\"") {
                curName = strings.Trim(curName, "\"")
            }
        case "test:fixture":
            kv := parseRuntimePayload(d.Payload)
            fx := amiFixture{path: kv["path"], mode: kv["mode"]}
            curFixtures = append(curFixtures, fx)
        case "test:runtime":
            name := curName
            if name == "" {
                name = filepath.Base(file)
            }
            rc := amiRuntimeCase{name: name, file: file, pkg: pkg}
            // Parse payload key=value pairs
            kv := parseRuntimePayload(d.Payload)
            rc.pipeline = kv["pipeline"]
            rc.inputJSON = kv["input"]
            rc.expectJSON = kv["expect_output"]
            rc.expectError = kv["expect_error"]
            if t := strings.TrimSpace(kv["timeout"]); t != "" {
                if n, err := strconv.Atoi(t); err == nil && n >= 0 {
                    rc.timeoutMs = n
                }
            }
            // attach fixtures snapshot
            if len(curFixtures) > 0 {
                rc.fixtures = append([]amiFixture{}, curFixtures...)
            }
            cases = append(cases, rc)
        }
    }
    return cases
}

