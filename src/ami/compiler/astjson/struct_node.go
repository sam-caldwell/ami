package astjson

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

func buildStructNode(n astpkg.StructDecl) sch.ASTNode {
    st := sch.ASTNode{Kind: "StructDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
    st.Fields = map[string]interface{}{"name": n.Name}
    return st
}

