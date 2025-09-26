package types

import "testing"

func TestFunction_String_Render(t *testing.T) {
    f := Function{Params: []Type{TInt, TString}, Results: []Type{TBool}}
    if got := f.String(); got != "(int,string) -> (bool)" {
        t.Fatalf("unexpected function string: %q", got)
    }
}

