package trigger

import (
    stdtime "time"
    amitime "github.com/sam-caldwell/ami/src/ami/stdlib/time"
)

// toAMI converts a stdlib time.Time to amitime.Time.
func toAMI(t stdtime.Time) amitime.Time { return amitime.FromUnix(t.Unix(), int64(t.Nanosecond())) }

