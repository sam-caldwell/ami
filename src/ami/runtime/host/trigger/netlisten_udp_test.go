package trigger

import (
    "net"
    "strconv"
    "testing"
    "time"
    amiio "github.com/sam-caldwell/ami/src/ami/runtime/host/io"
)

func TestNetListen_UDP_ReceivesAndMetadata(t *testing.T) {
    l, err := NetListen(amiio.UDP, "127.0.0.1", 0)
    if err != nil { t.Fatalf("NetListen UDP: %v", err) }
    defer l.Close()

    // Parse bound address
    host, portStr, err := net.SplitHostPort(l.LocalAddr())
    if err != nil { t.Fatalf("split host:port: %v", err) }
    p, _ := strconv.Atoi(portStr)

    // Send a datagram
    conn, err := net.Dial("udp", net.JoinHostPort(host, strconv.Itoa(p)))
    if err != nil { t.Fatalf("dial udp: %v", err) }
    defer conn.Close()
    if _, err := conn.Write([]byte("hello")); err != nil { t.Fatalf("send: %v", err) }

    select {
    case e := <-l.Events():
        if string(e.Value.Payload) != "hello" { t.Fatalf("payload=%q", string(e.Value.Payload)) }
        if e.Value.Protocol != amiio.UDP { t.Fatalf("protocol mismatch") }
        if e.Value.RemoteHost == "" || e.Value.LocalHost == "" { t.Fatalf("missing address metadata") }
    case <-time.After(1*time.Second):
        t.Fatalf("timeout waiting for udp event")
    }
}

