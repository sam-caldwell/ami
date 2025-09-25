package time

import stdtime "time"

// FormatDuration formats a duration string deterministically using Go's canonical form.
func FormatDuration(d stdtime.Duration) string { return d.String() }

