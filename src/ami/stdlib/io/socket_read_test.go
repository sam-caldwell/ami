package io

import (
	"net"
	"strconv"
	"testing"
	"time"
)

func testTCPSocket_Read_FromServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer ln.Close()
	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	p, _ := strconv.Atoi(portStr)

	// Accept and write response
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_, _ = c.Write([]byte("pong"))
	}()

	s, err := OpenSocket(TCP, host, uint16(p))
	if err != nil {
		t.Fatalf("OpenSocket TCP: %v", err)
	}
	defer s.Close()
	buf := make([]byte, 8)
	n, err := s.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(buf[:n]) != "pong" {
		t.Fatalf("got %q", string(buf[:n]))
	}
}

func testUDPSocket_Read_Datagram(t *testing.T) {
	srv, err := OpenSocket(UDP, "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("OpenSocket UDP: %v", err)
	}
	defer srv.Close()
	la := srv.LocalAddr()
	host, portStr, _ := net.SplitHostPort(la)
	p, _ := strconv.Atoi(portStr)

	// client send datagram
	cli, err := OpenSocket(UDP, "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("cli: %v", err)
	}
	defer cli.Close()
	if _, err := cli.Write([]byte("hello")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := cli.SendTo(host, uint16(p)); err != nil {
		t.Fatalf("sendto: %v", err)
	}

	// read on server
	buf := make([]byte, 16)
	_ = srv.pc.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, err := srv.Read(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(buf[:n]) != "hello" {
		t.Fatalf("got %q", string(buf[:n]))
	}
}

func testSocket_LocalRemoteAddr(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer ln.Close()
	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	p, _ := strconv.Atoi(portStr)

	// server accept and hold
	connected := make(chan struct{}, 1)
	go func() {
		c, err := ln.Accept()
		if err == nil {
			connected <- struct{}{}
			<-time.After(50 * time.Millisecond)
			_ = c.Close()
		}
	}()

	s, err := OpenSocket(TCP, host, uint16(p))
	if err != nil {
		t.Fatalf("OpenSocket TCP: %v", err)
	}
	defer s.Close()
	<-connected
	if s.LocalAddr() == "" {
		t.Fatalf("local addr empty")
	}
	if s.RemoteAddr() == "" {
		t.Fatalf("remote addr empty")
	}

	// UDP
	u, err := OpenSocket(UDP, "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("udp open: %v", err)
	}
	defer u.Close()
	if u.LocalAddr() == "" {
		t.Fatalf("udp local addr empty")
	}
	if u.RemoteAddr() != "" {
		t.Fatalf("udp remote should be empty")
	}
}
