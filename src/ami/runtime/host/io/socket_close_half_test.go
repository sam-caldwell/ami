package io

import (
	"net"
	"strconv"
	"testing"
	"time"
)

func testTCPCloseWrite_NoError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer ln.Close()
	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	p, _ := strconv.Atoi(portStr)

	// accept in background but ignore data path specifics
	go func() {
		c, _ := ln.Accept()
		if c != nil {
			_ = c.Close()
		}
	}()

	s, err := ConnectTCP(host, uint16(p))
	if err != nil {
		t.Fatalf("connect tcp: %v", err)
	}
	defer s.Close()
	if _, err := s.Write([]byte("abc")); err != nil {
		t.Fatalf("write: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if err := s.CloseWrite(); err != nil {
		t.Fatalf("CloseWrite: %v", err)
	}
}

func testTCPCloseRead_NoError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer ln.Close()
	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	p, _ := strconv.Atoi(portStr)

	go func() {
		c, _ := ln.Accept()
		if c != nil {
			_ = c.Close()
		}
	}()

	s, err := ConnectTCP(host, uint16(p))
	if err != nil {
		t.Fatalf("connect tcp: %v", err)
	}
	defer s.Close()
	if err := s.CloseRead(); err != nil {
		t.Fatalf("CloseRead: %v", err)
	}
}
