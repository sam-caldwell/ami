package io

import (
    "errors"
    "net"
    "strconv"
)

// Socket is a lightweight network handle with buffered writes.
// This initial implementation provides UDP binding and send/listen helpers.
type Socket struct {
    proto  NetProtocol
    pc     net.PacketConn // UDP bound socket
    conn   net.Conn       // TCP connected socket
    buf    []byte
    closed bool
}

// OpenSocket opens a UDP socket bound to addr:port. ICMP/TCP are not implemented yet.
func OpenSocket(proto NetProtocol, addr string, port uint16) (*Socket, error) {
    switch proto {
    case UDP:
        pc, err := net.ListenPacket("udp", net.JoinHostPort(addr, strconv.Itoa(int(port))))
        if err != nil { return nil, err }
        return &Socket{proto: UDP, pc: pc}, nil
    case TCP:
        c, err := net.Dial("tcp", net.JoinHostPort(addr, strconv.Itoa(int(port))))
        if err != nil { return nil, err }
        return &Socket{proto: TCP, conn: c}, nil
    case ICMP:
        return nil, errors.New("ICMP not implemented")
    default:
        return nil, errors.New("unknown protocol")
    }
}

// Write appends p to the internal buffer.
func (s *Socket) Write(p []byte) (int, error) {
    if s == nil || s.closed { return 0, ErrClosed }
    s.buf = append(s.buf, p...)
    return len(p), nil
}

// Send sends the buffered data for connected sockets. For UDP bound sockets, use SendTo.
func (s *Socket) Send() error {
    if s == nil || s.closed { return ErrClosed }
    // TCP connected path
    if s.conn != nil {
        if len(s.buf) == 0 { return nil }
        _, err := s.conn.Write(s.buf)
        s.buf = s.buf[:0]
        return err
    }
    // UDP listener path (no connected remote)
    return errors.New("Send not supported on UDP listening sockets; use SendTo")
}

// SendTo transmits the buffered data to the given remote host:port for UDP.
func (s *Socket) SendTo(host string, port uint16) error {
    if s == nil || s.closed { return ErrClosed }
    if s.proto != UDP || s.pc == nil { return errors.New("SendTo requires UDP listening socket") }
    if len(s.buf) == 0 { return nil }
    raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, strconv.Itoa(int(port))))
    if err != nil { return err }
    _, err = s.pc.WriteTo(s.buf, raddr)
    s.buf = s.buf[:0]
    return err
}

// Listen registers a handler that receives each incoming datagram on UDP sockets.
func (s *Socket) Listen(handler func(b []byte)) error {
    if s == nil || s.closed { return ErrClosed }
    switch s.proto {
    case UDP:
        if s.pc == nil { return errors.New("Listen requires UDP listening socket") }
        go func() {
            buf := make([]byte, 64*1024)
            for {
                n, _, err := s.pc.ReadFrom(buf)
                if err != nil { return }
                cp := make([]byte, n)
                copy(cp, buf[:n])
                handler(cp)
            }
        }()
        return nil
    case TCP:
        if s.conn == nil { return errors.New("Listen requires TCP connected socket") }
        go func() {
            buf := make([]byte, 64*1024)
            for {
                n, err := s.conn.Read(buf)
                if err != nil { return }
                cp := make([]byte, n)
                copy(cp, buf[:n])
                handler(cp)
            }
        }()
        return nil
    default:
        return errors.New("unsupported protocol")
    }
}

// Close closes the socket.
func (s *Socket) Close() error {
    if s == nil || s.closed { return nil }
    s.closed = true
    var err error
    if s.pc != nil { err = s.pc.Close() }
    if s.conn != nil { _ = s.conn.Close() }
    return err
}
