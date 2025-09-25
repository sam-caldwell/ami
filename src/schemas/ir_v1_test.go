package schemas

import (
	"encoding/json"
	"testing"
)

func TestIRV1_ValidateAndMarshal(t *testing.T) {
	ir := &IRV1{Schema: "ir.v1", Package: "p", File: "f", Functions: []IRFunction{{Name: "main", Blocks: []IRBlock{{Label: "entry"}}}}}
	if err := ir.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	b, err := json.Marshal(ir)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got IRV1
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Schema != "ir.v1" {
		t.Fatalf("unexpected schema: %s", got.Schema)
	}
}
