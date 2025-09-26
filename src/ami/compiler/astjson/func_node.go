package astjson

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

func buildFuncNode(n astpkg.FuncDecl) sch.ASTNode {
    fn := sch.ASTNode{Kind: "FuncDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
    fn.Fields = map[string]interface{}{"name": n.Name}
    return fn
}

