package time

import (
    stdtime "time"
    "testing"
)

func TestTime_FormatParseRFC3339Millis_RoundTrip(t *testing.T) {
    t0 := stdtime.Date(2025, 9, 25, 13, 45, 6, 123000000, stdtime.FixedZone("X", -7*3600))
    s := FormatRFC3339Millis(t0)
    if s != "2025-09-25T20:45:06.123Z" { t.Fatalf("format got %q", s) }
    back, err := ParseRFC3339Millis(s)
    if err != nil { t.Fatalf("parse error: %v", err) }
    if !back.Equal(t0.UTC().Truncate(stdtime.Millisecond)) { t.Fatalf("roundtrip mismatch: %v vs %v", back, t0) }
}

func TestTime_Duration_ParseFormat_RoundTrip(t *testing.T) {
    d, err := ParseDuration("1h2m3s4ms")
    if err != nil { t.Fatal(err) }
    if FormatDuration(d) != "1h2m3.004s" { t.Fatalf("format got %q", FormatDuration(d)) }
    d2, err := ParseDuration(FormatDuration(d))
    if err != nil { t.Fatal(err) }
    if d2 != d { t.Fatal("duration roundtrip mismatch") }
}

func TestTime_Clock_Fixed(t *testing.T) {
    now := stdtime.Date(2000,1,2,3,4,5,0,stdtime.UTC)
    c := FixedClock{T: now}
    if !c.Now().Equal(now) { t.Fatal("fixed clock mismatch") }
}

func TestTime_Sad_ParseError(t *testing.T) {
    if _, err := ParseRFC3339Millis("2025-09-25 12:00:00"); err == nil { t.Fatal("expected parse error") }
    if _, err := ParseDuration("notaduration"); err == nil { t.Fatal("expected duration parse error") }
}

