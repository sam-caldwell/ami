package driver

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// Verify that diagnostics can be emitted as JSON lines including file and pos.
func TestDiagnostics_JSONEmission_IncludesFileAndPos(t *testing.T) {
    fs := &source.FileSet{}
    // Create a pipeline missing ingress to force E_PIPELINE_START_INGRESS
    code := "package app\npipeline P(){ work; egress }\n"
    fs.AddFile("bad.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(workspace.Workspace{}, pkgs, Options{Debug: false})
    if len(diags) == 0 { t.Fatalf("expected diagnostics") }
    // Encode first diagnostic as JSON line
    b := diag.Line(diags[0])
    s := string(b)
    if !strings.Contains(s, "\"schema\":\"diag.v1\"") {
        t.Fatalf("missing schema in diag json: %s", s)
    }
    if !strings.Contains(s, "\"file\":\"bad.ami\"") {
        t.Fatalf("missing file field in diag json: %s", s)
    }
    if !strings.Contains(s, "\"pos\":{") {
        t.Fatalf("missing pos object in diag json: %s", s)
    }
}

// Ensure unresolved identifier diagnostics are emitted with consistent JSON fields.
func TestDiagnostics_JSON_UnresolvedIdent_IncludesFields(t *testing.T) {
    fs := &source.FileSet{}
    // unresolved ident 'y' in return
    code := "package app\nfunc F(){ return y }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(workspace.Workspace{}, pkgs, Options{Debug: false})
    if len(diags) == 0 {
        t.Fatalf("expected diagnostics for unresolved ident")
    }
    // Find E_UNRESOLVED_IDENT
    var d diag.Record
    for _, r := range diags { if r.Code == "E_UNRESOLVED_IDENT" { d = r; break } }
    if d.Code == "" { t.Fatalf("E_UNRESOLVED_IDENT not found: %+v", diags) }
    b := diag.Line(d)
    s := string(b)
    if !strings.Contains(s, "\"schema\":\"diag.v1\"") { t.Fatalf("missing schema: %s", s) }
    if !strings.Contains(s, "\"code\":\"E_UNRESOLVED_IDENT\"") { t.Fatalf("missing code: %s", s) }
    if !strings.Contains(s, "\"message\":\"unresolved identifier") { t.Fatalf("missing message: %s", s) }
    if !strings.Contains(s, "\"file\":\"u.ami\"") { t.Fatalf("missing file: %s", s) }
    if !strings.Contains(s, "\"pos\":{") { t.Fatalf("missing pos: %s", s) }
}

// no extra helpers
