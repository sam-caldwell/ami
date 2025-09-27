package diag

import (
    "encoding/json"
    "strings"
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
    if err != nil {
        t.Fatalf("marshal: %v", err)
    }
    s := string(b)
    want := `{"schema":"diag.v1","timestamp":"2025-09-24T17:05:06.123Z","level":"error","code":"E1001","message":"unexpected token \";\"","package":"example/app","file":"pkg/main.ami","pos":{"line":12,"column":17,"offset":214},"data":{"a":1,"b":2}}`
    if s != want {
        t.Fatalf("unexpected json order/body:\nwant: %s\n got: %s", want, s)
    }
}

func TestDiagRecord_OptionalFieldsOmitted(t *testing.T) {
    r := Record{Timestamp: time.Unix(0, 0), Level: Info, Code: "I000", Message: "ok"}
    b, _ := json.Marshal(r)
    s := string(b)
    if strings.Contains(s, "package") || strings.Contains(s, "file") || strings.Contains(s, "pos") || strings.Contains(s, "endPos") || strings.Contains(s, "data") {
        t.Fatalf("optional fields should not appear when empty: %s", s)
    }
}

func TestDiagLine_AppendsNewline(t *testing.T) {
    r := Record{Timestamp: time.Unix(0, 0), Level: Warn, Code: "W001", Message: "warn"}
    line := Line(r)
    if !strings.HasSuffix(string(line), "\n") {
        t.Fatalf("expected newline suffix")
    }
}
