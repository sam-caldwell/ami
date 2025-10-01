package io

import (
    "net"
    "strconv"
    "testing"
    "time"
)

func TestTCPSocket_Connect_Write_Send(t *testing.T) {
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatalf("listen tcp: %v", err) }
    defer ln.Close()
    host, portStr, _ := net.SplitHostPort(ln.Addr().String())
    p, _ := strconv.Atoi(portStr)

    accepted := make(chan []byte, 1)
    go func(){
        c, err := ln.Accept(); if err != nil { return }
        defer c.Close()
        buf := make([]byte, 16)
        n, _ := c.Read(buf)
        accepted <- append([]byte{}, buf[:n]...)
    }()

    s, err := OpenSocket(TCP, host, uint16(p))
    if err != nil { t.Fatalf("OpenSocket TCP: %v", err) }
    defer s.Close()
    if _, err := s.Write([]byte("abc")); err != nil { t.Fatalf("write: %v", err) }
    if err := s.Send(); err != nil { t.Fatalf("send: %v", err) }

    select {
    case b := <-accepted:
        if string(b) != "abc" { t.Fatalf("server got %q want 'abc'", string(b)) }
    case <-time.After(1 * time.Second):
        t.Fatalf("timeout waiting for server receive")
    }
}

func TestTCPSocket_Connect_Listen_Receives(t *testing.T) {
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatalf("listen tcp: %v", err) }
    defer ln.Close()
    host, portStr, _ := net.SplitHostPort(ln.Addr().String())
    p, _ := strconv.Atoi(portStr)

    // Accept and write a message to client
    go func(){
        c, err := ln.Accept(); if err != nil { return }
        defer c.Close()
        _, _ = c.Write([]byte("pong"))
    }()

    s, err := OpenSocket(TCP, host, uint16(p))
    if err != nil { t.Fatalf("OpenSocket TCP: %v", err) }
    defer s.Close()

    recv := make(chan []byte, 1)
    if err := s.Listen(func(b []byte){ recv <- append([]byte{}, b...) }); err != nil { t.Fatalf("listen: %v", err) }

    select {
    case b := <-recv:
        if string(b) != "pong" { t.Fatalf("client got %q want 'pong'", string(b)) }
    case <-time.After(1 * time.Second):
        t.Fatalf("timeout waiting for client receive")
    }
}

