package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerFuncDecl lowers a single function declaration into IR with a single entry block.
func lowerFuncDeclWithSCC(fn *ast.FuncDecl, funcResMap, funcParamMap map[string][]string, funcParamNames map[string][]string, sccSet map[string]bool) ir.Function {
    var params []ir.Value
    var outResults []ir.Value
    st := &lowerState{varTypes: map[string]string{}, funcResults: funcResMap, funcParams: funcParamMap, funcParamNames: funcParamNames, currentFn: fn.Name, methodRecv: map[string]irValue{}}
    for _, p := range fn.Params {
        params = append(params, ir.Value{ID: p.Name, Type: p.Type})
        if p.Name != "" && p.Type != "" { st.varTypes[p.Name] = p.Type }
    }
    for _, r := range fn.Results {
        outResults = append(outResults, ir.Value{Type: r.Type})
    }
    instrs, extraBlocks := lowerBlockCFG(st, fn.Body, 0)
    // collect decorators
    var decos []ir.Decorator
    for _, d := range fn.Decorators {
        var args []string
        for _, e := range d.Args { args = append(args, debugExprText(e)) }
        decos = append(decos, ir.Decorator{Name: d.Name, Args: args})
    }
    // Tail-call elimination (self/mutual) M19: if last return is preceded by a call expr
    if len(instrs) >= 2 {
        if ret, ok := instrs[len(instrs)-1].(ir.Return); ok {
            if ex, ok2 := instrs[len(instrs)-2].(ir.Expr); ok2 && ex.Op == "call" && ex.Callee == fn.Name {
                // Replace the call+return with loop markers
                instrs = instrs[:len(instrs)-2]
                instrs = append(instrs, ir.Loop{Name: fn.Name})
                instrs = append(instrs, ir.Goto{Label: "entry"})
                // Synthesize a default return to keep emitter happy
                instrs = append(instrs, ret)
            } else if ok2 && ex.Op == "call" && sccSet != nil && sccSet[ex.Callee] {
                // Mutual recursion tail-call: mark dispatch and goto
                instrs = instrs[:len(instrs)-2]
                instrs = append(instrs, ir.Loop{Name: fn.Name})
                instrs = append(instrs, ir.Dispatch{Label: ex.Callee})
                instrs = append(instrs, ir.Goto{Label: "entry"})
                instrs = append(instrs, ret)
            }
        }
    }
    blocks := []ir.Block{{Name: "entry", Instr: instrs}}
    if len(extraBlocks) > 0 { blocks = append(blocks, extraBlocks...) }
    // Map collected GPU blocks
    var gbs []ir.GPUBlock
    for _, g := range st.gpuBlocks {
        gb := ir.GPUBlock{Family: g.family, Name: g.name, Source: g.source, N: g.n, Grid: g.grid, TPG: g.tpg, Args: g.args}
        gbs = append(gbs, gb)
    }
    return ir.Function{Name: fn.Name, Params: params, Results: outResults, Blocks: blocks, Decorators: decos, GPUBlocks: gbs}
}
// helper moved to lower_func_debugtext.go
