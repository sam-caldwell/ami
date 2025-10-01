//go:build runtime_posix && (darwin || linux || freebsd)

package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Ensure extern for ami_rt_posix_install_trampoline is declared when used.
func TestEmit_Extern_PosixInstall_Trampoline(t *testing.T) {
    m := ir.Module{Package: "app"}
    f := ir.Function{Name: "F"}
    f.Blocks = append(f.Blocks, ir.Block{Name: "entry", Instr: []ir.Instruction{
        ir.Expr{Op: "call", Callee: "ami_rt_posix_install_trampoline", Args: []ir.Value{{ID: "#2", Type: "int64"}}},
        ir.Return{},
    }})
    m.Functions = append(m.Functions, f)
    s, err := EmitModuleLLVMForTarget(m, DefaultTriple)
    if err != nil { t.Fatalf("emit: %v", err) }
    if !strings.Contains(s, "declare void @ami_rt_posix_install_trampoline(i64)") {
        t.Fatalf("missing extern declaration for posix install:\n%s", s)
    }
    if !strings.Contains(s, "call void @ami_rt_posix_install_trampoline(i64 2)") {
        t.Fatalf("missing call for posix install:\n%s", s)
    }
}

