package io

import (
	"net"
	"strconv"
	"testing"
	"time"
)

func testUDPSocket_ReadDeadline_Timeout(t *testing.T) {
	s, err := ListenUDP("127.0.0.1", 0)
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer s.Close()
	if err := s.SetReadDeadline(time.Now().Add(20 * time.Millisecond)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
	buf := make([]byte, 4)
	_, err = s.Read(buf)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if ne, ok := err.(net.Error); !ok || !ne.Timeout() {
		t.Fatalf("expected net.Error timeout, got %T %v", err, err)
	}
}

func testTCPSocket_ReadDeadline_Timeout(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer ln.Close()
	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	p, _ := strconv.Atoi(portStr)
	// Accept and hold without writing
	done := make(chan struct{}, 1)
	go func() {
		c, _ := ln.Accept()
		<-time.After(100 * time.Millisecond)
		if c != nil {
			_ = c.Close()
		}
		done <- struct{}{}
	}()
	s, err := ConnectTCP(host, uint16(p))
	if err != nil {
		t.Fatalf("connect tcp: %v", err)
	}
	defer s.Close()
	if err := s.SetReadDeadline(time.Now().Add(20 * time.Millisecond)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
	buf := make([]byte, 4)
	_, err = s.Read(buf)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if ne, ok := err.(net.Error); !ok || !ne.Timeout() {
		t.Fatalf("expected timeout, got %T %v", err, err)
	}
	<-done
}
