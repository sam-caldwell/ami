package schemas

import (
	"encoding/json"
	"testing"
)

func TestObjIndexV1_ValidateAndMarshal(t *testing.T) {
	idx := ObjIndexV1{Schema: "objindex.v1", Package: "p", Files: []ObjFile{{Unit: "u", Path: "build/obj/p/u.s", Size: 1, Sha256: "deadbeef"}}}
	if err := idx.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	b, err := json.Marshal(idx)
	if err != nil || len(b) == 0 {
		t.Fatalf("marshal err=%v", err)
	}
	var got ObjIndexV1
	if json.Unmarshal(b, &got) != nil {
		t.Fatalf("unmarshal")
	}
	if got.Schema != "objindex.v1" || got.Package != "p" || len(got.Files) != 1 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}
