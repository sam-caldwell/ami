package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestEmitModuleLLVM_EmbedsErrorPipelineGlobals(t *testing.T) {
    m := ir.Module{Package: "app", ErrorPipes: []ir.ErrorPipeline{{Pipeline: "P", Steps: []string{"ingress", "Transform", "egress"}}}}
    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(out, "@ami_errpipe_P = private constant") {
        t.Fatalf("missing error pipeline global: %s", out)
    }
    if !strings.Contains(out, "pipeline:P|steps:ingress,Transform,egress\\00") {
        t.Fatalf("payload malformed or missing: %s", out)
    }
}

