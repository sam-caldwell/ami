package root

import (
    "testing"
    diagpkg "github.com/sam-caldwell/ami/src/ami/compiler/diag"
)

func TestEvalAmiExpectations_NoErrors_NoWarnings(t *testing.T) {
    diags := []diagpkg.Diagnostic{}
    c := amiCase{name: "noerr", pkg: "p", expects: []amiExpect{{kind: "no_errors"}, {kind: "no_warnings"}}}
    ok, fail := evalAmiExpectations(c, diags)
    if !ok || len(fail) != 0 { t.Fatalf("expected pass; got ok=%v fail=%v", ok, fail) }
}

func TestEvalAmiExpectations_ErrorCodeCountAndMsgSubstr(t *testing.T) {
    diags := []diagpkg.Diagnostic{
        {Level: diagpkg.Error, Code: "E_BAD_IMPORT", Message: "invalid import path: bad"},
        {Level: diagpkg.Error, Code: "E_BAD_IMPORT", Message: "bad import path again"},
        {Level: diagpkg.Warn, Code: "W_FILE_NO_NEWLINE", Message: "file missing newline"},
    }
    // Expect exactly two E_BAD_IMPORT
    c1 := amiCase{name: "errs2", pkg: "p", expects: []amiExpect{{kind: "error", code: "E_BAD_IMPORT", countSet: true, count: 2}}}
    ok, _ := evalAmiExpectations(c1, diags)
    if !ok { t.Fatalf("expected pass for count=2") }
    // Expect one error with substring 'again'
    c2 := amiCase{name: "substr", pkg: "p", expects: []amiExpect{{kind: "error", code: "E_BAD_IMPORT", countSet: true, count: 1, msgSubstr: "again"}}}
    ok, _ = evalAmiExpectations(c2, diags)
    if !ok { t.Fatalf("expected pass for substring match") }
    // Expect total errors count == 2 and warnings count == 1
    c3 := amiCase{name: "agg", pkg: "p", expects: []amiExpect{{kind: "errors_count", countSet: true, count: 2}, {kind: "warnings_count", countSet: true, count: 1}}}
    ok, _ = evalAmiExpectations(c3, diags)
    if !ok { t.Fatalf("expected pass for aggregate counts") }
}

