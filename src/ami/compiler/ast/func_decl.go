package ast

import tok "github.com/sam-caldwell/ami/src/ami/compiler/token"

// FuncDecl represents a function declaration with parameters, results, and body.
// Body is captured as tokens for scaffolded semantic checks; BodyStmts holds a
// minimal statement AST parsed from those tokens.
type FuncDecl struct {
    Name      string
    TypeParams []TypeParam
    Params    []Param
    Result    []TypeRef
    Body      []tok.Token
    BodyStmts []Stmt
    Pos       Position
    Comments  []Comment
}

func (FuncDecl) isNode() {}
