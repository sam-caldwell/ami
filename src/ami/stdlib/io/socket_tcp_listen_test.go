package io

import (
    "net"
    "strconv"
    "testing"
    "time"
)

func TestTCPServer_ListenSocket_AcceptsAndReads(t *testing.T) {
    srv, err := ListenSocket(TCP, "127.0.0.1", 0)
    if err != nil { t.Fatalf("ListenSocket TCP: %v", err) }
    defer srv.Close()

    la := srv.ln.Addr().String()
    host, portStr, _ := net.SplitHostPort(la)
    p, _ := strconv.Atoi(portStr)

    recv := make(chan []byte, 1)
    if err := srv.Listen(func(b []byte){ recv <- append([]byte{}, b...) }); err != nil { t.Fatalf("Listen: %v", err) }

    // Client connects and writes
    c, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(p)))
    if err != nil { t.Fatalf("dial: %v", err) }
    defer c.Close()
    if _, err := c.Write([]byte("hello")); err != nil { t.Fatalf("client write: %v", err) }

    select {
    case b := <-recv:
        if string(b) != "hello" { t.Fatalf("server received %q", string(b)) }
    case <-time.After(1 * time.Second):
        t.Fatalf("timeout waiting for server receive")
    }
}

func TestICMPOpen_NotImplemented(t *testing.T) {
    if _, err := OpenSocket(ICMP, "127.0.0.1", 0); err == nil {
        t.Fatalf("expected ICMP not implemented error")
    }
    if _, err := ListenSocket(ICMP, "127.0.0.1", 0); err == nil {
        t.Fatalf("expected ICMP not implemented error (listen)")
    }
}

