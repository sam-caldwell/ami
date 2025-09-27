package ir

import "testing"

func TestModule_FunctionsAppendAndAccess(t *testing.T) {
    m := Module{Package: "main"}
    if m.Package != "main" || len(m.Functions) != 0 { t.Fatalf("unexpected module: %+v", m) }
    f := Function{Name: "F", Params: []Value{{ID: "p0", Type: "int"}}, Results: []Value{{ID: "r0", Type: "int"}}}
    m.Functions = append(m.Functions, f)
    if len(m.Functions) != 1 || m.Functions[0].Name != "F" { t.Fatalf("unexpected functions: %+v", m.Functions) }
}

func TestFunction_BlocksAndInstrTypes(t *testing.T) {
    b := Block{Name: "entry"}
    b.Instr = append(b.Instr, Var{}, Assign{}, Return{}, Defer{}, Expr{})
    if len(b.Instr) != 5 { t.Fatalf("unexpected instr len: %d", len(b.Instr)) }
    fn := Function{Name: "F", Blocks: []Block{b}}
    if fn.Blocks[0].Name != "entry" { t.Fatalf("block name") }
}

