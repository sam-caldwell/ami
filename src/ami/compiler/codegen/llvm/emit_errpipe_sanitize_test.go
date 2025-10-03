package llvm

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestEmitModuleLLVM_ErrorPipeGlobalName_Sanitized(t *testing.T) {
    m := ir.Module{Package: "app", ErrorPipes: []ir.ErrorPipeline{{Pipeline: "P-1:name", Steps: []string{"ingress","egress"}}}}
    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    if indexOf(out, "@ami_errpipe_P_1_name = private constant") < 0 {
        t.Fatalf("sanitized name missing: %s", out)
    }
}

