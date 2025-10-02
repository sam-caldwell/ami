package io

import (
    "errors"
    "net"
    "strconv"
    "testing"
    "time"
    "os"
)

func TestReadWriteBytes_Aliases(t *testing.T) {
    f, err := CreateTemp()
    if err != nil { t.Fatalf("CreateTemp: %v", err) }
    defer func(){ _ = f.Close() }()
    if n, err := f.WriteBytes([]byte("abc")); err != nil || n != 3 { t.Fatalf("WriteBytes: %d %v", n, err) }
    if _, err := f.Seek(0, 0); err != nil { t.Fatalf("Seek: %v", err) }
    buf := make([]byte, 3)
    if n, err := f.ReadBytes(buf); err != nil || n != 3 || string(buf) != "abc" { t.Fatalf("ReadBytes: %d %v %q", n, err, string(buf)) }
}

func TestCreateTemp_InvalidArgs_Error(t *testing.T) {
    if _, err := CreateTemp("a", "b", "c"); err == nil {
        t.Fatalf("expected error on invalid CreateTemp args")
    }
}

func TestName_AfterClose_AndNil(t *testing.T) {
    // Nil receiver -> empty
    var nilF *FHO
    if nilF.Name() != "" { t.Fatalf("nil Name should be empty") }
    f, err := CreateTemp()
    if err != nil { t.Fatalf("CreateTemp: %v", err) }
    name := f.Name()
    if err := f.Close(); err != nil { t.Fatalf("close: %v", err) }
    if f.Name() != name { t.Fatalf("Name should remain after close") }
}

func TestHostname_CapabilityDenied(t *testing.T) {
    defer ResetPolicy()
    SetPolicy(Policy{AllowFS:true, AllowNet:false, AllowDevice:true})
    if _, err := Hostname(); err == nil || !errors.Is(err, ErrCapabilityDenied) {
        t.Fatalf("expected capability denied for Hostname; got %v", err)
    }
}

func TestName_Fallback_FromFileDescriptor(t *testing.T) {
    h := &FHO{f: os.Stdout, name: ""}
    if h.Name() == "" { t.Fatalf("expected fallback name from file descriptor") }
}

func TestConnectSocket_UDP_ReturnsError(t *testing.T) {
    if _, err := ConnectSocket(UDP, "127.0.0.1", 0); err == nil {
        t.Fatalf("expected UDP ConnectSocket error")
    }
}

func TestListenAndOpenSocket_UnknownProtocol_Error(t *testing.T) {
    bad := NetProtocol("BAD")
    if _, err := OpenSocket(bad, "127.0.0.1", 0); err == nil {
        t.Fatalf("expected OpenSocket unknown protocol error")
    }
    if _, err := ListenSocket(bad, "127.0.0.1", 0); err == nil {
        t.Fatalf("expected ListenSocket unknown protocol error")
    }
    if _, err := ConnectSocket(bad, "127.0.0.1", 0); err == nil {
        t.Fatalf("expected ConnectSocket unknown protocol error")
    }
    if _, err := ConnectSocket(ICMP, "127.0.0.1", 0); err == nil {
        t.Fatalf("expected ConnectSocket ICMP error")
    }
}

func TestSendTo_EmptyBuffer_NoError(t *testing.T) {
    s, err := ListenUDP("127.0.0.1", 0)
    if err != nil { t.Fatalf("ListenUDP: %v", err) }
    defer s.Close()
    // No writes -> empty internal buffer; SendTo should be a no-op
    la := s.LocalAddr()
    host, portStr, _ := net.SplitHostPort(la)
    p, _ := strconv.Atoi(portStr)
    if err := s.SendTo(host, uint16(p)); err != nil { t.Fatalf("SendTo empty: %v", err) }
}

func TestReadFrom_UDP_ReturnsHostPort(t *testing.T) {
    srv, err := ListenUDP("127.0.0.1", 0)
    if err != nil { t.Fatalf("ListenUDP: %v", err) }
    defer srv.Close()
    la := srv.LocalAddr()
    host, portStr, _ := net.SplitHostPort(la)
    p, _ := strconv.Atoi(portStr)

    cli, err := ListenUDP("127.0.0.1", 0)
    if err != nil { t.Fatalf("client ListenUDP: %v", err) }
    defer cli.Close()
    if _, err := cli.Write([]byte("xyz")); err != nil { t.Fatalf("write: %v", err) }
    if err := cli.SendTo(host, uint16(p)); err != nil { t.Fatalf("sendto: %v", err) }

    buf := make([]byte, 8)
    if n, rhost, rport, err := srv.ReadFrom(buf); err != nil || n != 3 || rhost == "" || rport == 0 {
        t.Fatalf("ReadFrom: n=%d err=%v host=%q port=%d", n, err, rhost, rport)
    }
}

type dummyListener struct{}
func (d dummyListener) Accept() (net.Conn, error) { time.Sleep(5 * time.Millisecond); return nil, errors.New("done") }
func (d dummyListener) Close() error               { return nil }
func (d dummyListener) Addr() net.Addr             { return &net.IPAddr{} }

func TestListener_Deadlines_WhenNotTCPListener(t *testing.T) {
    s := &Socket{proto: TCP, ln: dummyListener{}}
    if err := s.SetDeadline(time.Now()); err == nil { t.Fatalf("expected SetDeadline error on dummy listener") }
    if err := s.SetReadDeadline(time.Now()); err == nil { t.Fatalf("expected SetReadDeadline error on dummy listener") }
    if err := s.SetWriteDeadline(time.Now()); err == nil { t.Fatalf("expected SetWriteDeadline error on listener") }
}

type notTCPConn struct{ net.Conn }

func TestCloseReadWrite_NotTCPConn_Error(t *testing.T) {
    s := &Socket{proto: TCP, conn: notTCPConn{Conn: nil}}
    if err := s.CloseRead(); err == nil { t.Fatalf("CloseRead should error for non-TCPConn") }
    if err := s.CloseWrite(); err == nil { t.Fatalf("CloseWrite should error for non-TCPConn") }
}

func TestCreateTempDir_And_GetPolicy(t *testing.T) {
    dir, err := CreateTempDir()
    if err != nil || dir == "" { t.Fatalf("CreateTempDir: %v %q", err, dir) }
    p := GetPolicy()
    if !p.AllowFS || !p.AllowNet || !p.AllowDevice { t.Fatalf("unexpected default policy: %+v", p) }
}

func TestWriteTo_UDP_Sends(t *testing.T) {
    srv, err := ListenUDP("127.0.0.1", 0)
    if err != nil { t.Fatalf("ListenUDP: %v", err) }
    defer srv.Close()
    // Client
    cli, err := ListenUDP("127.0.0.1", 0)
    if err != nil { t.Fatalf("cli ListenUDP: %v", err) }
    defer cli.Close()

    // Determine client address
    chost, cportStr, _ := net.SplitHostPort(cli.LocalAddr())
    cp, _ := strconv.Atoi(cportStr)

    // Server writes to client
    if n, err := srv.WriteTo(chost, uint16(cp), []byte("hi")); err != nil || n != 2 { t.Fatalf("WriteTo: %d %v", n, err) }

    // Client receives
    buf := make([]byte, 8)
    if n, rhost, rport, err := cli.ReadFrom(buf); err != nil || n != 2 || rhost == "" || rport == 0 { t.Fatalf("cli.ReadFrom: %d %v %q %d", n, err, rhost, rport) }
}

func TestReadFrom_TCP_ReturnsZeros(t *testing.T) {
    // Create a TCP server that writes then closes
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatalf("listen: %v", err) }
    defer ln.Close()
    host, portStr, _ := net.SplitHostPort(ln.Addr().String())
    p, _ := strconv.Atoi(portStr)
    done := make(chan struct{}, 1)
    go func(){ c, _ := ln.Accept(); if c!=nil { _, _ = c.Write([]byte("ok")); _ = c.Close() }; close(done) }()
    s, err := ConnectTCP(host, uint16(p))
    if err != nil { t.Fatalf("connect: %v", err) }
    defer s.Close()
    buf := make([]byte, 8)
    n, rhost, rport, err := s.ReadFrom(buf)
    if err != nil || n != 2 || rhost != "" || rport != 0 {
        t.Fatalf("ReadFrom TCP: n=%d err=%v host=%q port=%d", n, err, rhost, rport)
    }
    <-done
}

func TestSetWriteDeadline_TCP(t *testing.T) {
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatalf("listen: %v", err) }
    defer ln.Close()
    host, portStr, _ := net.SplitHostPort(ln.Addr().String())
    p, _ := strconv.Atoi(portStr)
    go func(){ c, _ := ln.Accept(); if c!=nil { _ = c.Close() } }()
    s, err := ConnectTCP(host, uint16(p))
    if err != nil { t.Fatalf("connect: %v", err) }
    defer s.Close()
    if err := s.SetWriteDeadline(time.Now()); err != nil { t.Fatalf("SetWriteDeadline TCP: %v", err) }
}

func TestSendTo_WrongSocketType_Error(t *testing.T) {
    s := &Socket{proto: TCP}
    if err := s.SendTo("127.0.0.1", 9); err == nil { t.Fatalf("expected SendTo error for TCP socket") }
}

func TestRead_NoUnderlyingSocket_Error(t *testing.T) {
    s := &Socket{}
    buf := make([]byte, 1)
    if _, err := s.Read(buf); err == nil { t.Fatalf("expected error for Read with no underlying socket") }
}

func TestServe_NoListener_Error(t *testing.T) {
    s := &Socket{proto: TCP}
    if err := s.Serve(func(*Socket){}); err == nil { t.Fatalf("expected error for Serve without listener") }
}

func TestOpenFile_RDWR_Create(t *testing.T) {
    f, err := CreateTemp()
    if err != nil { t.Fatalf("CreateTemp: %v", err) }
    name := f.Name()
    _ = f.Close()
    h, err := OpenFile(name, os.O_RDWR|os.O_CREATE, 0o644)
    if err != nil { t.Fatalf("OpenFile rdwr: %v", err) }
    defer h.Close()
}

func TestClosedHandle_Errors(t *testing.T) {
    f, err := CreateTemp()
    if err != nil { t.Fatalf("CreateTemp: %v", err) }
    _ = f.Close()
    if _, err := f.Pos(); err == nil { t.Fatalf("expected Pos error on closed handle") }
    if _, err := f.Length(); err == nil { t.Fatalf("expected Length error on closed handle") }
    if err := f.Truncate(0); err == nil { t.Fatalf("expected Truncate error on closed handle") }
    if err := f.Flush(); err == nil { t.Fatalf("expected Flush error on closed handle") }
}
