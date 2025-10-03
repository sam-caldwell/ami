package exec

import (
    "testing"
    stdtime "time"
    amitime "github.com/sam-caldwell/ami/src/ami/stdlib/time"
)

func Test_toStdTime_RoundtripFields(t *testing.T) {
    // Construct a Time via Now and reconstruct expected from Unix/UnixNano
    at := amitime.Now()
    sec := at.Unix()
    nsec := at.UnixNano() - sec*1_000_000_000
    want := stdtime.Unix(sec, nsec).UTC()
    got := toStdTime(at)
    if !got.Equal(want) { t.Fatalf("toStdTime mismatch: got=%v want=%v", got, want) }
}

