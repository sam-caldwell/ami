package llvm

import (
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Ensure externs are declared for handler thunk install/get when referenced.
func TestEmit_HandlerThunk_Externs(t *testing.T) {
    m := ir.Module{Package: "app"}
    f := ir.Function{Name: "F"}
    // call void @ami_rt_install_handler_thunk(i64 <tok>, ptr %fp)
    f.Blocks = append(f.Blocks, ir.Block{Name: "entry", Instr: []ir.Instruction{
        ir.Expr{Op: "call", Callee: "ami_rt_install_handler_thunk", Args: []ir.Value{{ID: "#42", Type: "int64"}, {ID: "fp", Type: "ptr"}}},
        // %p = call ptr @ami_rt_get_handler_thunk(i64 <tok>)
        ir.Expr{Op: "call", Callee: "ami_rt_get_handler_thunk", Args: []ir.Value{{ID: "#42", Type: "int64"}}, Result: &ir.Value{ID: "p", Type: "ptr"}},
    }})
    m.Functions = append(m.Functions, f)
    out, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    wants := []string{
        "declare void @ami_rt_install_handler_thunk(i64, ptr)",
        "declare ptr @ami_rt_get_handler_thunk(i64)",
        "call void @ami_rt_install_handler_thunk(",
        "call ptr @ami_rt_get_handler_thunk(",
    }
    for _, w := range wants {
        if !strings.Contains(out, w) {
            t.Fatalf("missing fragment: %q\n%s", w, out)
        }
    }
}

