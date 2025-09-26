package codegen

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestGenerateASM_Golden(t *testing.T) {
    m := ir.Module{Package: "p", Unit: "u.ami", Functions: []ir.Function{{Name: "a"}, {Name: "b"}}}
    got := GenerateASM(m)
    want := "; AMI-IR assembly\n; package p unit u.ami\nfn_a:\n  ret\nfn_b:\n  ret\n"
    if got != want {
        t.Fatalf("\n--- got ---\n%q\n--- want ---\n%q", got, want)
    }
}

