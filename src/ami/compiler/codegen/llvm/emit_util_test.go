package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func Test_sanitizeIdent(t *testing.T) {
    if got := sanitizeIdent("P-1:name"); got != "P_1_name" { t.Fatalf("sanitize: %q", got) }
    if got := sanitizeIdent(""); got != "_" { t.Fatalf("empty sanitize: %q", got) }
}

func Test_itoa(t *testing.T) {
    if itoa(0) != "0" || itoa(12345) != "12345" { t.Fatalf("itoa failed") }
}

func Test_buildModuleMetaJSON(t *testing.T) {
    m := ir.Module{Package: "pkg", Concurrency: 2, Backpressure: "block", Schedule: "fair", TelemetryEnabled: true, Capabilities: []string{"io"}, TrustLevel: "trusted"}
    m.Pipelines = []ir.Pipeline{{Name: "P"}}
    m.ErrorPipes = []ir.ErrorPipeline{{Pipeline: "P", Steps: []string{"ingress","egress"}}}
    s := buildModuleMetaJSON(m)
    if !strings.Contains(s, "\"schema\":\"ami.meta.v1\"") { t.Fatalf("schema: %s", s) }
    if !strings.Contains(s, "\"package\":\"pkg\"") { t.Fatalf("package: %s", s) }
    if !strings.Contains(s, "\"concurrency\":2") { t.Fatalf("concurrency: %s", s) }
    if !strings.Contains(s, "\"pipelines\":") { t.Fatalf("pipelines: %s", s) }
    if !strings.Contains(s, "\"errorPipelines\":") { t.Fatalf("errpipes: %s", s) }
}

