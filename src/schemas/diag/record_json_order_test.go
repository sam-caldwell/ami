package diag

import (
    "encoding/json"
    "testing"
    "time"
)

func TestDiagRecord_JSONOrderingAndFields(t *testing.T) {
    r := Record{
        Timestamp: time.Date(2025, 9, 24, 17, 5, 6, 123000000, time.UTC),
        Level:     Error,
        Code:      "E1001",
        Message:   "unexpected token \";\"",
        Package:   "example/app",
        File:      "pkg/main.ami",
        Pos:       &Position{Line: 12, Column: 17, Offset: 214},
        Data:      map[string]any{"b": 2, "a": 1},
    }
    b, err := json.Marshal(r)
    if err != nil { t.Fatalf("marshal: %v", err) }
    s := string(b)
    want := `{"schema":"diag.v1","timestamp":"2025-09-24T17:05:06.123Z","level":"error","code":"E1001","message":"unexpected token \";\"","package":"example/app","file":"pkg/main.ami","pos":{"line":12,"column":17,"offset":214},"data":{"a":1,"b":2}}`
    if s != want { t.Fatalf("unexpected json order/body:\nwant: %s\n got: %s", want, s) }
}

