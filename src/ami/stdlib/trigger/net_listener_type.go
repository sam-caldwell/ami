package trigger

import (
    "sync"
    amiio "github.com/sam-caldwell/ami/src/ami/stdlib/io"
)

// NetListener encapsulates network event emission for supported protocols.
type NetListener struct {
    proto amiio.NetProtocol
    sock  *amiio.Socket
    out   chan Event[NetMsg]
    wg    sync.WaitGroup
    once  sync.Once
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

