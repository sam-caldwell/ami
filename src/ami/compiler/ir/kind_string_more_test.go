package ir

import "testing"

func TestKind_String_All(t *testing.T) {
    kinds := []Kind{OpVar, OpAssign, OpReturn, OpDefer, OpExpr, OpPhi, OpCondBr, OpLoop, OpGoto, OpSetPC, OpDispatch, OpPushFrame, OpPopFrame, Kind(999)}
    for _, k := range kinds {
        s := k.String()
        if s == "" { t.Fatalf("empty for %v", k) }
    }
}

