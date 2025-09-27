package logging

import (
    "bytes"
    "testing"
    "time"
)

func TestJSONFormatter_EmitsLogV1(t *testing.T) {
    f := JSONFormatter{}
    r := Record{Timestamp: time.Unix(0,0), Level: LevelInfo, Message: "hi", Package: "pkg", Fields: map[string]any{"k":1}}
    b := f.Format(r)
    if !bytes.Contains(b, []byte("\"schema\":\"log.v1\"")) {
        t.Fatalf("missing schema: %s", string(b))
    }
}

