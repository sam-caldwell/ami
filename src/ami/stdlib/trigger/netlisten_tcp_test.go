package trigger

import (
    "net"
    "strconv"
    "testing"
    "time"

    amiio "github.com/sam-caldwell/ami/src/ami/stdlib/io"
)

func TestNetListen_TCP_ReceivesAndMetadata(t *testing.T) {
    l, err := NetListen(amiio.TCP, "127.0.0.1", 0)
    if err != nil { t.Fatalf("NetListen TCP: %v", err) }
    defer l.Close()

    // Determine actual port
    host, portStr, err := net.SplitHostPort(l.LocalAddr())
    if err != nil { t.Fatalf("split host:port: %v", err) }
    p, _ := strconv.Atoi(portStr)

    // Send a message from a client
    c, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(p)))
    if err != nil { t.Fatalf("dial: %v", err) }
    defer c.Close()
    if _, err := c.Write([]byte("hello")); err != nil { t.Fatalf("client write: %v", err) }

    select {
    case e := <-l.Events():
        if string(e.Value.Payload) != "hello" {
            t.Fatalf("payload=%q", string(e.Value.Payload))
        }
        if e.Value.Protocol != amiio.TCP { t.Fatalf("protocol mismatch") }
        if e.Value.RemoteHost == "" || e.Value.LocalHost == "" {
            t.Fatalf("missing addressing metadata: %+v", e.Value)
        }
        if e.Value.RemotePort == 0 || e.Value.LocalPort == 0 {
            t.Fatalf("missing ports in metadata: %+v", e.Value)
        }
    case <-time.After(1 * time.Second):
        t.Fatalf("timeout waiting for network event")
    }
}

