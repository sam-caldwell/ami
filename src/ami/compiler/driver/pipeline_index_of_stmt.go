package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

func indexOfStmt(stmts []ast.Stmt, target ast.Stmt) int {
    for i, s := range stmts { if s == target { return i } }
    return -1
}

