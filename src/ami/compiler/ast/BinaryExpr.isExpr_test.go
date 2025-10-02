package ast

import "testing"

func Test_BinaryExpr_isExpr(t *testing.T) {
	e := BinaryExpr{}
	e.isExpr()
}
