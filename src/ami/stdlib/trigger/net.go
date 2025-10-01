package trigger

import (
    "net"
    "strconv"
    "sync"

    amitime "github.com/sam-caldwell/ami/src/ami/stdlib/time"
    amiio "github.com/sam-caldwell/ami/src/ami/stdlib/io"
)

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

// NetListener encapsulates network event emission for supported protocols.
type NetListener struct {
    proto amiio.NetProtocol
    sock  *amiio.Socket
    out   chan Event[NetMsg]
    wg    sync.WaitGroup
    once  sync.Once
}

// NetListen starts listening on addr:port for the given protocol and emits NetMsg events.
// Currently, TCP is supported. UDP/ICMP require additional io features (remote addr delivery).
func NetListen(proto amiio.NetProtocol, addr string, port uint16) (*NetListener, error) {
    switch proto {
    case amiio.TCP:
        s, err := amiio.ListenSocket(amiio.TCP, addr, port)
        if err != nil { return nil, err }
        l := &NetListener{proto: amiio.TCP, sock: s, out: make(chan Event[NetMsg], 1024)}
        // Accept/serve and spawn per-connection readers.
        if err := s.Serve(func(cs *amiio.Socket) {
            l.wg.Add(1)
            go func(c *amiio.Socket) {
                defer l.wg.Done()
                defer c.Close()
                buf := make([]byte, 64*1024)
                for {
                    n, err := c.Read(buf)
                    if err != nil { return }
                    if n == 0 { continue }
                    rh, rp := splitHostPort(c.RemoteAddr())
                    lh, lp := splitHostPort(c.LocalAddr())
                    payload := make([]byte, n)
                    copy(payload, buf[:n])
                    l.out <- Event[NetMsg]{
                        Value: NetMsg{
                            Protocol:   amiio.TCP,
                            Payload:    payload,
                            RemoteHost: rh, RemotePort: rp,
                            LocalHost:  lh, LocalPort:  lp,
                            Time:       amitime.Now(),
                        },
                        Timestamp: amitime.Now(),
                    }
                }
            }(cs)
        }); err != nil {
            _ = s.Close()
            return nil, err
        }
        return l, nil
    case amiio.UDP:
        return nil, ErrNotImplemented
    case amiio.ICMP:
        return nil, ErrNotImplemented
    default:
        return nil, ErrNotImplemented
    }
}

// Events returns the receive-only channel for NetMsg events.
func (l *NetListener) Events() <-chan Event[NetMsg] { return l.out }

// LocalAddr returns the local address string of the underlying listener.
func (l *NetListener) LocalAddr() string { if l == nil || l.sock == nil { return "" }; return l.sock.LocalAddr() }

// Close stops the listener and closes the events channel after readers exit.
func (l *NetListener) Close() error {
    if l == nil || l.sock == nil { return nil }
    var err error
    l.once.Do(func(){ err = l.sock.Close(); l.wg.Wait(); close(l.out) })
    return err
}

// splitHostPort attempts to split a host:port string into parts.
func splitHostPort(addr string) (string, uint16) {
    if addr == "" { return "", 0 }
    host, portStr, err := net.SplitHostPort(addr)
    if err != nil { return addr, 0 }
    p, _ := strconv.Atoi(portStr)
    if p < 0 { p = 0 }
    if p > 65535 { p = 65535 }
    return host, uint16(p)
}

