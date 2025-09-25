package ast

import "testing"

func TestAST_Scaffold_TypesSatisfyInterfaces(t *testing.T) {
    var _ Node = ImportDecl{}
    var _ Node = FuncDecl{}
    var _ Node = PipelineDecl{}
    var _ Node = Bad{}
    var _ Node = EdgeSpec{}
    var _ Node = PackageDecl{}

    var _ Expr = Ident{}
    var _ Expr = BasicLit{}
    var _ Expr = CallExpr{}
}

