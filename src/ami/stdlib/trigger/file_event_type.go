package trigger

import amitime "github.com/sam-caldwell/ami/src/ami/stdlib/time"
import amiio "github.com/sam-caldwell/ami/src/ami/stdlib/io"

// FileEvent is the payload for filesystem notifications.
// Handle may be nil for events where a handle is unavailable (e.g., removed).
type FileEvent struct {
    Handle *amiio.FHO
    Op     FsEvent
    Time   amitime.Time
}

