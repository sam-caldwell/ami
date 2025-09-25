package schemas

import (
    "encoding/json"
    "testing"
)

func TestASTV1_ValidateAndMarshal(t *testing.T) {
    a := &ASTV1{Schema:"ast.v1", Package:"p", File:"f", Root: ASTNode{Kind:"File", Pos: Position{Line:1,Column:1,Offset:0}}}
    if err := a.Validate(); err != nil { t.Fatalf("validate: %v", err) }
    b, err := json.Marshal(a)
    if err != nil { t.Fatalf("marshal: %v", err) }
    var got ASTV1
    if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("unmarshal: %v", err) }
    if got.Schema != "ast.v1" { t.Fatalf("unexpected schema: %s", got.Schema) }
}

