package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "github.com/spf13/cobra"
    codegen "github.com/sam-caldwell/ami/src/ami/compiler/codegen"
    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)


func newEmitWrappersCmd() *cobra.Command {
    var out string
    var triple string
    var specsPath string
    cmd := &cobra.Command{
        Use:    "emit-wrappers",
        Short:  "Emit LLVM worker core wrappers only (hidden)",
        Hidden: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            if out == "" || specsPath == "" {
                return cmd.Help()
            }
            if triple == "" { triple = llvme.DefaultTriple }
            // read specs
            b, err := os.ReadFile(specsPath)
            if err != nil { return err }
            var specs []wrapperSpec
            if err := json.Unmarshal(b, &specs); err != nil { return err }
            // build IR module with worker-shaped signatures
            var fns []ir.Function
            for _, s := range specs {
                if s.Name == "" || s.Param == "" || s.Result == "" { continue }
                f := ir.Function{ Name: s.Name, Params: []ir.Value{{ID: "ev", Type: s.Param}}, Results: []ir.Value{{ID: "r0", Type: s.Result}, {ID: "r1", Type: "error"}} }
                fns = append(fns, f)
            }
            m := ir.Module{Package: "app", Functions: fns}
            ll, err := codegen.EmitWorkerWrappersOnlyForTarget(m, triple)
            if err != nil { return err }
            if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil { return err }
            return os.WriteFile(out, []byte(ll), 0o644)
        },
    }
    cmd.Flags().StringVar(&out, "out", "", "output .ll path")
    cmd.Flags().StringVar(&triple, "triple", "", "target triple (default: host)")
    cmd.Flags().StringVar(&specsPath, "specs", "", "specs JSON: [{name,param,result}...]")
    return cmd
}
