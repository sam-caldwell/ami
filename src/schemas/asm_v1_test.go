package schemas

import (
	"encoding/json"
	"testing"
)

func TestASMIndexV1_ValidateAndMarshal(t *testing.T) {
	idx := &ASMIndexV1{Schema: "asm.v1", Package: "p", Files: []ASMFile{{Unit: "u", Path: "build/debug/asm/u.s", Size: 1, Sha256: "deadbeef"}}}
	if err := idx.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	b, err := json.Marshal(idx)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got ASMIndexV1
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Schema != "asm.v1" {
		t.Fatalf("unexpected schema: %s", got.Schema)
	}
}
