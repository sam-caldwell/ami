package ast

import "testing"

func Test_ErrorBlock_isNode(t *testing.T) {
	e := ErrorBlock{}
	e.isNode()
}
