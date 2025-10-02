package ast

import "testing"

func Test_BinaryExpr_isNode(t *testing.T) {
	e := BinaryExpr{}
	e.isNode()
}
