package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerFuncDecl lowers a single function declaration into IR with a single entry block.
func lowerFuncDecl(fn *ast.FuncDecl) ir.Function {
    var params []ir.Value
    var results []ir.Value
    for _, p := range fn.Params {
        params = append(params, ir.Value{ID: p.Name, Type: p.Type})
    }
    for _, r := range fn.Results {
        results = append(results, ir.Value{Type: r.Type})
    }
    st := &lowerState{}
    instrs := lowerBlock(st, fn.Body)
    blk := ir.Block{Name: "entry", Instr: instrs}
    return ir.Function{Name: fn.Name, Params: params, Results: results, Blocks: []ir.Block{blk}}
}

