package schemas

import (
    "encoding/json"
    "testing"
    "time"
)

func TestDiagV1_MarshalUnmarshal_Validate(t *testing.T) {
    d := &DiagV1{
        Schema:    "diag.v1",
        Timestamp: FormatTimestamp(time.Unix(0, 0)),
        Level:     "error",
        Code:      "E1001",
        Message:   "unexpected token",
        File:      "pkg/main.ami",
        Pos:       &Position{Line: 1, Column: 1, Offset: 0},
    }
    if err := d.Validate(); err != nil {
        t.Fatalf("validate failed: %v", err)
    }
    b, err := json.Marshal(d)
    if err != nil {
        t.Fatalf("marshal: %v", err)
    }
    var got DiagV1
    if err := json.Unmarshal(b, &got); err != nil {
        t.Fatalf("unmarshal: %v", err)
    }
    if got.Schema != "diag.v1" || got.Level != "error" || got.Message == "" {
        t.Fatalf("unexpected decoded diag: %+v", got)
    }
}

