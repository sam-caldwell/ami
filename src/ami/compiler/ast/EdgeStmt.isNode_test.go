package ast

import "testing"

func Test_EdgeStmt_isNode(t *testing.T) {
	e := EdgeStmt{}
	e.isNode()
}
