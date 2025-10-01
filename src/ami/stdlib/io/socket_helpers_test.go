package io

import (
    "net"
    "strconv"
    "testing"
    "time"
)

func TestConnectSocket_TCP_Works_UDP_Fails(t *testing.T) {
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatalf("listen tcp: %v", err) }
    defer ln.Close()
    host, portStr, _ := net.SplitHostPort(ln.Addr().String())
    p, _ := strconv.Atoi(portStr)

    // Accept and close
    done := make(chan struct{}, 1)
    go func(){ c, _ := ln.Accept(); if c != nil { _ = c.Close() }; done<-struct{}{} }()

    s, err := ConnectSocket(TCP, host, uint16(p))
    if err != nil { t.Fatalf("ConnectSocket TCP: %v", err) }
    s.Close()
    select { case <-done: case <-time.After(200*time.Millisecond): t.Fatalf("server no accept") }

    if _, err := ConnectSocket(UDP, "127.0.0.1", 0); err == nil { t.Fatalf("expected udp connect error") }
}

func TestWriteTo_UDP(t *testing.T) {
    srv, err := ListenUDP("127.0.0.1", 0)
    if err != nil { t.Fatalf("udp listen: %v", err) }
    defer srv.Close()
    host, portStr, _ := net.SplitHostPort(srv.LocalAddr())
    p, _ := strconv.Atoi(portStr)

    // Read in background
    recv := make(chan []byte, 1)
    go func(){ buf := make([]byte, 16); n, _ := srv.Read(buf); recv<-append([]byte{}, buf[:n]...) }()

    cli, err := ListenUDP("127.0.0.1", 0)
    if err != nil { t.Fatalf("udp client: %v", err) }
    defer cli.Close()
    if n, err := cli.WriteTo(host, uint16(p), []byte("xyz")); err != nil || n != 3 { t.Fatalf("WriteTo: n=%d err=%v", n, err) }

    select {
    case b := <-recv:
        if string(b) != "xyz" { t.Fatalf("got %q", string(b)) }
    case <-time.After(1 * time.Second):
        t.Fatalf("timeout waiting for udp read")
    }
}

func TestServe_TCP_AcceptsAndHandlerCanWrite(t *testing.T) {
    srv, err := ListenTCP("127.0.0.1", 0)
    if err != nil { t.Fatalf("listen tcp: %v", err) }
    defer srv.Close()
    if err := srv.Serve(func(c *Socket){ _, _ = c.Write([]byte("hi")); _ = c.Send(); _ = c.Close() }); err != nil { t.Fatalf("serve: %v", err) }

    host, portStr, _ := net.SplitHostPort(srv.LocalAddr())
    p, _ := strconv.Atoi(portStr)
    cli, err := ConnectTCP(host, uint16(p))
    if err != nil { t.Fatalf("connect tcp: %v", err) }
    defer cli.Close()
    buf := make([]byte, 8)
    n, err := cli.Read(buf)
    if err != nil { t.Fatalf("read: %v", err) }
    if string(buf[:n]) != "hi" { t.Fatalf("got %q", string(buf[:n])) }
}

