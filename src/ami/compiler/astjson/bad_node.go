package astjson

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

func buildBadNode(n astpkg.Bad) sch.ASTNode {
    bad := sch.ASTNode{Kind: "Bad", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
    bad.Fields = map[string]interface{}{"token": n.Tok.Kind, "lexeme": n.Tok.Lexeme}
    return bad
}

