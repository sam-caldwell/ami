package ast

// VarDeclStmt declares a local variable with optional type and initializer.
type VarDeclStmt struct {
    Name     string
    Type     TypeRef
    Init     Expr
    Pos      Position
    Comments []Comment
}

func (VarDeclStmt) isStmt() {}

