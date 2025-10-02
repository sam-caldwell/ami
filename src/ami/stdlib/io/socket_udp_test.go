package io

import (
	"net"
	"sync"
	"testing"
	"time"
)

func testUDPSocket_Open_Write_SendTo_Listen_Close(t *testing.T) {
	s, err := OpenSocket(UDP, "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("OpenSocket UDP: %v", err)
	}
	defer s.Close()

	pc := s.pc
	if pc == nil {
		t.Fatalf("expected PacketConn")
	}
	la := pc.LocalAddr().(*net.UDPAddr)

	got := make(chan []byte, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	if err := s.Listen(func(b []byte) { got <- b; wg.Done() }); err != nil {
		t.Fatalf("Listen: %v", err)
	}

	client, err := OpenSocket(UDP, "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("client OpenSocket: %v", err)
	}
	defer client.Close()
	if _, err := client.Write([]byte("ping")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := client.SendTo("127.0.0.1", uint16(la.Port)); err != nil {
		t.Fatalf("sendto: %v", err)
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
		select {
		case b := <-got:
			if string(b) != "ping" {
				t.Fatalf("got %q", string(b))
			}
		default:
			t.Fatalf("no msg")
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for udp recv")
	}

	// Close idempotent
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close 2: %v", err)
	}
}

func testUDPSocket_Send_UnsupportedAndClosedErrors(t *testing.T) {
	s, err := OpenSocket(UDP, "127.0.0.1", 0)
	if err != nil {
		t.Fatalf("OpenSocket UDP: %v", err)
	}
	if _, err := s.Write([]byte("x")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := s.Send(); err == nil {
		t.Fatalf("expected error for Send on UDP listener")
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := s.Write([]byte("y")); err == nil || err != ErrClosed {
		t.Fatalf("expected ErrClosed write, got %v", err)
	}
	if err := s.Send(); err == nil || err != ErrClosed {
		t.Fatalf("expected ErrClosed send, got %v", err)
	}
}
