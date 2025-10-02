package trigger

import amitime "github.com/sam-caldwell/ami/src/ami/stdlib/time"
import amiio "github.com/sam-caldwell/ami/src/ami/stdlib/io"

// NetMsg represents a received network message with addressing metadata.
type NetMsg struct {
    Protocol   amiio.NetProtocol
    Payload    []byte
    RemoteHost string
    RemotePort uint16
    LocalHost  string
    LocalPort  uint16
    Time       amitime.Time
}

