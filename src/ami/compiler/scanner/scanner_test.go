package scanner

import (
    "reflect"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// TestScanner_StructDefinition verifies the structure and field types of Scanner.
func TestScanner_StructDefinition(t *testing.T) {
    typ := reflect.TypeOf(Scanner{})

    if typ.Kind() != reflect.Struct {
        t.Fatalf("Scanner should be a struct; got %v", typ.Kind())
    }

    if typ.NumField() != 2 {
        t.Fatalf("Scanner should have 2 fields; got %d", typ.NumField())
    }

    // Field 0: file *source.File (unexported)
    f0 := typ.Field(0)
    if f0.Name != "file" {
        t.Fatalf("field[0] name mismatch; want 'file', got %q", f0.Name)
    }
    wantFileType := reflect.TypeOf((*source.File)(nil))
    if f0.Type != wantFileType {
        t.Fatalf("field[0] type mismatch; want %v, got %v", wantFileType, f0.Type)
    }
    if f0.PkgPath == "" { // empty means exported
        t.Fatalf("field[0] should be unexported")
    }

    // Field 1: offset int (unexported)
    f1 := typ.Field(1)
    if f1.Name != "offset" {
        t.Fatalf("field[1] name mismatch; want 'offset', got %q", f1.Name)
    }
    if f1.Type.Kind() != reflect.Int {
        t.Fatalf("field[1] type mismatch; want int, got %v", f1.Type)
    }
    if f1.PkgPath == "" { // empty means exported
        t.Fatalf("field[1] should be unexported")
    }
}
