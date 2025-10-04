package trigger

import (
    amitime "github.com/sam-caldwell/ami/src/ami/runtime/host/time"
    amiio "github.com/sam-caldwell/ami/src/ami/runtime/host/io"
)

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
        s, err := amiio.ListenSocket(amiio.UDP, addr, port)
        if err != nil { return nil, err }
        l := &NetListener{proto: amiio.UDP, sock: s, out: make(chan Event[NetMsg], 1024)}
        // Spin a reader for datagrams capturing remote address metadata.
        l.wg.Add(1)
        go func(){
            defer l.wg.Done()
            buf := make([]byte, 64*1024)
            for {
                n, rh, rp, err := s.ReadFrom(buf)
                if err != nil { return }
                if n == 0 { continue }
                lh, lp := splitHostPort(s.LocalAddr())
                payload := make([]byte, n)
                copy(payload, buf[:n])
                l.out <- Event[NetMsg]{
                    Value: NetMsg{
                        Protocol:   amiio.UDP,
                        Payload:    payload,
                        RemoteHost: rh, RemotePort: rp,
                        LocalHost:  lh, LocalPort:  lp,
                        Time:       amitime.Now(),
                    },
                    Timestamp: amitime.Now(),
                }
            }
        }()
        return l, nil
    case amiio.ICMP:
        return nil, ErrNotImplemented
    default:
        return nil, ErrNotImplemented
    }
}

