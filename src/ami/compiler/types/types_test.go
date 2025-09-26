package types

import "testing"

func TestBasic_String(t *testing.T) {
    if TInt.String() != "int" || TString.String() != "string" || TBool.String() != "bool" {
        t.Fatalf("basic type names: int=%q string=%q bool=%q", TInt, TString, TBool)
    }
}

