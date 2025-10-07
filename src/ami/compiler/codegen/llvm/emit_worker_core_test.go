package llvm

import (
    "strings"
    "testing"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestEmit_WorkerCore_WrapperPresent(t *testing.T) {
    // Define a worker-shaped function: func W(ev Event<T>) (Event<U>, error)
    fn := ir.Function{Name: "W", Params: []ir.Value{{ID: "ev", Type: "Event<int>"}}, Results: []ir.Value{{ID: "r0", Type: "Event<int>"}, {ID: "r1", Type: "error"}}}
    m := ir.Module{Package: "app", Functions: []ir.Function{fn}}
    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    // Check that wrapper is present
    if !strings.Contains(out, "define i8* @ami_worker_core_W(i8* %in, i32 %inlen, i32* %outlen, i8** %err)") {
        t.Fatalf("missing worker core wrapper:\n%s", out)
    }
}

