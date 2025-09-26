package astjson

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

func buildDirectiveNode(d astpkg.Directive) sch.ASTNode {
    dn := sch.ASTNode{Kind: "Directive", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
    dn.Fields = map[string]interface{}{"name": d.Name, "payload": d.Payload}
    return dn
}

