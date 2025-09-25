package types

import "testing"

func TestEventAndErrorTypes_String(t *testing.T) {
    e := EventType{Elem: TString}
    if e.String() != "Event<string>" { t.Fatalf("event string=%q", e.String()) }
    er := ErrorType{Elem: TInt}
    if er.String() != "Error<int>" { t.Fatalf("error string=%q", er.String()) }
}
