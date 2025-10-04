package exec

import (
    "time"
    amitime "github.com/sam-caldwell/ami/src/ami/runtime/host/time"
)

// toStdTime converts amitime.Time to stdlib time.Time for payload compatibility.
func toStdTime(t amitime.Time) time.Time {
    sec := t.Unix()
    nsec := t.UnixNano() - sec*1_000_000_000
    return time.Unix(sec, nsec).UTC()
}

