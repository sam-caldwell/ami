package driver

import "testing"

func TestLowerState_ZeroAndSet(t *testing.T) {
    var s lowerState
    if s.temp != 0 || s.varTypes != nil || s.funcResults != nil || s.funcParams != nil || s.funcParamNames != nil {
        t.Fatalf("unexpected zero value: %+v", s)
    }
    s.temp = 1
    s.varTypes = map[string]string{"x":"int"}
    s.funcResults = map[string][]string{"F": {"int"}}
    s.funcParams = map[string][]string{"F": {"int"}}
    s.funcParamNames = map[string][]string{"F": {"a"}}
    if s.temp != 1 || s.varTypes["x"] != "int" || len(s.funcResults["F"]) != 1 || s.funcParamNames["F"][0] != "a" {
        t.Fatalf("unexpected values: %+v", s)
    }
}

