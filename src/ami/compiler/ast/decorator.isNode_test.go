package ast

import "testing"

func TestDecorator_isNode(t *testing.T) {
	d := Decorator{}
	d.isNode()
}
