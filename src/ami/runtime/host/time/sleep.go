package amitime

import stdtime "time"

// Sleep pauses the current goroutine for at least duration d.
func Sleep(d Duration) { stdtime.Sleep(d) }

