package ast

import "testing"

func Test_EdgeStmt_isStmt(t *testing.T) {
	e := EdgeStmt{}
	e.isStmt()
}
