package root

import "time"

// nowUTC returns current time in UTC; extracted for testability.
func nowUTC() time.Time { return time.Now().UTC() }

