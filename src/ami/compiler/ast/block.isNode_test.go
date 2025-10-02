package ast

import "testing"

func TestBlock_isNode(t *testing.T) {
	block := BlockStmt{}
	block.isNode()
}
