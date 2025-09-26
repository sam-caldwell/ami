package astjson

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

func buildEnumNode(n astpkg.EnumDecl) sch.ASTNode {
    en := sch.ASTNode{Kind: "EnumDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
    en.Fields = map[string]interface{}{"name": n.Name}
    return en
}

