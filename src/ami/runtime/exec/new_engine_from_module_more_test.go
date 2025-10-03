package exec

import (
    "testing"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestNewEngineFromModule_Defaults(t *testing.T) {
    eng, err := NewEngineFromModule(ir.Module{})
    if err != nil || eng == nil || eng.pool == nil { t.Fatalf("engine: %v %+v", err, eng) }
    eng.Close() // exercise Close
}

func TestNewEngineFromModule_CustomPolicyWorkers(t *testing.T) {
    m := ir.Module{Schedule: "lifo", Concurrency: 2}
    eng, err := NewEngineFromModule(m)
    if err != nil || eng == nil || eng.pool == nil { t.Fatalf("engine: %v %+v", err, eng) }
    eng.Close()
}

