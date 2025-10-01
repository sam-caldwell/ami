package io

import (
    "testing"
    "time"
)

func TestSocket_SetDeadline_OnTCPListener(t *testing.T) {
    s, err := ListenTCP("127.0.0.1", 0)
    if err != nil { t.Fatalf("ListenTCP: %v", err) }
    defer s.Close()
    if err := s.SetDeadline(time.Now().Add(10 * time.Millisecond)); err != nil {
        t.Fatalf("SetDeadline on listener: %v", err)
    }
    if err := s.SetReadDeadline(time.Now().Add(10 * time.Millisecond)); err != nil {
        t.Fatalf("SetReadDeadline on listener: %v", err)
    }
    if err := s.SetWriteDeadline(time.Now()); err == nil {
        t.Fatalf("SetWriteDeadline should error for listener")
    }
}

func TestSocket_CloseReadCloseWrite_OnUDP_ReturnError(t *testing.T) {
    s, err := ListenUDP("127.0.0.1", 0)
    if err != nil { t.Fatalf("ListenUDP: %v", err) }
    defer s.Close()
    if err := s.CloseRead(); err == nil { t.Fatalf("CloseRead should error on UDP") }
    if err := s.CloseWrite(); err == nil { t.Fatalf("CloseWrite should error on UDP") }
}

func TestUDPSocket_SetWriteDeadline_OK(t *testing.T) {
    s, err := ListenUDP("127.0.0.1", 0)
    if err != nil { t.Fatalf("ListenUDP: %v", err) }
    defer s.Close()
    if err := s.SetWriteDeadline(time.Now().Add(5 * time.Millisecond)); err != nil {
        t.Fatalf("SetWriteDeadline on UDP: %v", err)
    }
}

func TestSocket_NoUnderlying_Errors(t *testing.T) {
    s := &Socket{}
    if err := s.SetDeadline(time.Now()); err == nil { t.Fatalf("expected error SetDeadline") }
    if err := s.SetReadDeadline(time.Now()); err == nil { t.Fatalf("expected error SetReadDeadline") }
    if err := s.SetWriteDeadline(time.Now()); err == nil { t.Fatalf("expected error SetWriteDeadline") }
}

