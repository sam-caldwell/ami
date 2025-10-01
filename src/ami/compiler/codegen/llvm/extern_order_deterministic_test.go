package llvm

import (
    "sort"
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// extractDecls returns the sorted list of extern declare lines from a module string.
func extractDecls(s string) []string {
    var decls []string
    for _, line := range strings.Split(s, "\n") {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "declare ") { decls = append(decls, line) }
        if strings.HasPrefix(line, "define ") { break }
    }
    out := append([]string(nil), decls...)
    sort.Strings(out)
    return out
}

// Ensure extern ordering is deterministic regardless of discovery order in IR traversal.
func TestModule_Externs_DeterministicOrder(t *testing.T) {
    // Module A: discover panic -> alloc -> zeroize
    a := ir.Module{Package: "p"}
    ab := ir.Block{Name: "entry"}
    ab.Instr = append(ab.Instr,
        ir.Expr{Op: "panic", Args: []ir.Value{{ID: "code", Type: "int32"}}},
        ir.Expr{Op: "call", Callee: "ami_rt_alloc", Args: []ir.Value{{ID: "sz", Type: "int64"}}},
        ir.Expr{Op: "call", Callee: "ami_rt_zeroize", Args: []ir.Value{{ID: "p", Type: "ptr"}, {ID: "n", Type: "int64"}}},
        ir.Return{},
    )
    a.Functions = []ir.Function{{Name: "F", Blocks: []ir.Block{ab}}}
    outA, err := EmitModuleLLVM(a)
    if err != nil { t.Fatalf("emit a: %v", err) }

    // Module B: discover zeroize -> alloc -> panic (reverse)
    b := ir.Module{Package: "p"}
    bb := ir.Block{Name: "entry"}
    bb.Instr = append(bb.Instr,
        ir.Expr{Op: "call", Callee: "ami_rt_zeroize", Args: []ir.Value{{ID: "p", Type: "ptr"}, {ID: "n", Type: "int64"}}},
        ir.Expr{Op: "call", Callee: "ami_rt_alloc", Args: []ir.Value{{ID: "sz", Type: "int64"}}},
        ir.Expr{Op: "panic", Args: []ir.Value{{ID: "code", Type: "int32"}}},
        ir.Return{},
    )
    b.Functions = []ir.Function{{Name: "F", Blocks: []ir.Block{bb}}}
    outB, err := EmitModuleLLVM(b)
    if err != nil { t.Fatalf("emit b: %v", err) }

    // Extract and compare sorted decls from both modules (should be identical)
    da := extractDecls(outA)
    db := extractDecls(outB)
    if len(da) == 0 || len(db) == 0 { t.Fatalf("no decls extracted: A=%v B=%v", da, db) }
    if len(da) != len(db) { t.Fatalf("decl count mismatch: %d vs %d\nA:%v\nB:%v", len(da), len(db), da, db) }
    for i := range da {
        if da[i] != db[i] {
            t.Fatalf("extern ordering not deterministic:\nA:%v\nB:%v", da, db)
        }
    }
}

