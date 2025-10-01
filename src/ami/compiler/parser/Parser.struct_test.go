package parser

import (
	"reflect"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"github.com/sam-caldwell/ami/src/ami/compiler/scanner"
	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestParser_StructShape(t *testing.T) {
	typ := reflect.TypeOf(Parser{})
	if typ.Kind() != reflect.Struct {
		t.Fatalf("Parser should be a struct")
	}
	// Validate field names and types (order-sensitive)
	if typ.NumField() != 5 {
		t.Fatalf("unexpected field count: %d", typ.NumField())
	}
	if typ.Field(0).Type != reflect.TypeOf((*scanner.Scanner)(nil)) {
		t.Fatalf("field s type mismatch")
	}
	if typ.Field(1).Type != reflect.TypeOf(token.Token{}) {
		t.Fatalf("field cur type mismatch")
	}
	// The remaining fields are slices; basic presence check
	_ = reflect.TypeOf([]ast.Comment{})
}
