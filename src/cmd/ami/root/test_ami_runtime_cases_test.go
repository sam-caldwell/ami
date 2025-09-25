package root

import (
    "testing"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// Verify deriveAmiRuntimeCases parses key fields from runtime pragma.
func Test_DeriveAmiRuntimeCases_Parse(t *testing.T) {
    f := &astpkg.File{Package: "p", Directives: []astpkg.Directive{
        {Name: "test:case", Payload: "Run1"},
        {Name: "test:runtime", Payload: "pipeline=Pipe input={\"k\":1} expect_output={\"k\":1} timeout=100"},
    }}
    cases := deriveAmiRuntimeCases("/x/y/runtime_test.ami", "p", f)
    if len(cases) != 1 { t.Fatalf("want 1 runtime case, got %d", len(cases)) }
    c := cases[0]
    if c.name != "Run1" { t.Fatalf("bad name: %q", c.name) }
    if c.pipeline != "Pipe" { t.Fatalf("bad pipeline: %q", c.pipeline) }
    if c.inputJSON != "{\"k\":1}" { t.Fatalf("bad input: %q", c.inputJSON) }
    if c.expectJSON != "{\"k\":1}" { t.Fatalf("bad expect: %q", c.expectJSON) }
    if c.timeoutMs != 100 { t.Fatalf("bad timeout: %d", c.timeoutMs) }
}

// Verify defaulting of case name to file basename when omitted.
func Test_DeriveAmiRuntimeCases_DefaultName(t *testing.T) {
    f := &astpkg.File{Package: "p", Directives: []astpkg.Directive{
        {Name: "test:runtime", Payload: "pipeline=P input={}"},
    }}
    cases := deriveAmiRuntimeCases("/x/y/z_test.ami", "p", f)
    if len(cases) != 1 { t.Fatalf("want 1 runtime case, got %d", len(cases)) }
    if cases[0].name != "z_test.ami" { t.Fatalf("expected default name z_test.ami, got %q", cases[0].name) }
}

