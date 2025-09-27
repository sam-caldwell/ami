package diag

import (
    "encoding/json"
    "testing"
)

func TestPosition_BasenamePair_JSON(t *testing.T) {
    b, err := json.Marshal(Position{Line:1, Column:2, Offset:3})
    if err != nil || string(b) == "" { t.Fatalf("marshal: %v", err) }
}

