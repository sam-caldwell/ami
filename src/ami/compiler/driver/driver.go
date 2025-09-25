package driver

import (
    sch "github.com/sam-caldwell/ami/src/schemas"
)

// Options placeholder for future flags
type Options struct{}

// Result holds compiler outputs for scaffolding
type Result struct {
    AST []sch.ASTV1
    IR  []sch.IRV1
    ASM []string // paths of asm files written by caller
}

// Compile is a placeholder; it constructs deterministic AST/IR from inputs
func Compile(files []string, opts Options) (Result, error) {
    res := Result{}
    for _, f := range files {
        ast := sch.ASTV1{Schema: "ast.v1", Package: "main", File: f, Root: sch.ASTNode{Kind: "File", Pos: sch.Position{Line:1,Column:1,Offset:0}}}
        res.AST = append(res.AST, ast)
        ir := sch.IRV1{Schema: "ir.v1", Package: "main", File: f, Functions: []sch.IRFunction{{Name:"main", Blocks: []sch.IRBlock{{Label:"entry"}}}}}
        res.IR = append(res.IR, ir)
    }
    return res, nil
}
