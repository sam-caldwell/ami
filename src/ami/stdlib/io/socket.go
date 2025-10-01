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
    ln     net.Listener   // TCP listening socket
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

// ConnectSocket is an alias for establishing a connected socket when supported.
// Currently supports TCP. For UDP, returns an error (use OpenSocket + SendTo).
func ConnectSocket(proto NetProtocol, host string, port uint16) (*Socket, error) {
    switch proto {
    case TCP:
        return OpenSocket(TCP, host, port)
    case UDP:
        return nil, errors.New("UDP ConnectSocket not supported; use OpenSocket + SendTo")
    case ICMP:
        return nil, errors.New("ICMP not implemented")
    default:
        return nil, errors.New("unknown protocol")
    }
}

// Convenience wrappers for common patterns.
func ConnectTCP(host string, port uint16) (*Socket, error)  { return ConnectSocket(TCP, host, port) }
func ListenTCP(addr string, port uint16) (*Socket, error)   { return ListenSocket(TCP, addr, port) }
func ListenUDP(addr string, port uint16) (*Socket, error)   { return OpenSocket(UDP, addr, port) }

// ListenSocket creates a listening socket for the given protocol.
// For UDP, this is equivalent to OpenSocket(UDP, addr, port).
// For TCP, it binds and listens on addr:port for incoming connections.
func ListenSocket(proto NetProtocol, addr string, port uint16) (*Socket, error) {
    switch proto {
    case UDP:
        return OpenSocket(UDP, addr, port)
    case TCP:
        ln, err := net.Listen("tcp", net.JoinHostPort(addr, strconv.Itoa(int(port))))
        if err != nil { return nil, err }
        return &Socket{proto: TCP, ln: ln}, nil
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
        // TCP server listener
        if s.ln != nil {
            go func() {
                for {
                    c, err := s.ln.Accept()
                    if err != nil { return }
                    go func(conn net.Conn){
                        defer conn.Close()
                        buf := make([]byte, 64*1024)
                        for {
                            n, err := conn.Read(buf)
                            if err != nil { return }
                            cp := make([]byte, n)
                            copy(cp, buf[:n])
                            handler(cp)
                        }
                    }(c)
                }
            }()
            return nil
        }
        // TCP connected socket
        if s.conn != nil {
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
        }
        return errors.New("Listen requires TCP connected or listening socket")
    default:
        return errors.New("unsupported protocol")
    }
}

// WriteTo sends p immediately to host:port on UDP listeners (non-buffered).
func (s *Socket) WriteTo(host string, port uint16, p []byte) (int, error) {
    if s == nil || s.closed { return 0, ErrClosed }
    if s.proto != UDP || s.pc == nil { return 0, errors.New("WriteTo requires UDP listening socket") }
    raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, strconv.Itoa(int(port))))
    if err != nil { return 0, err }
    return s.pc.WriteTo(p, raddr)
}

// Serve accepts TCP connections and invokes handler with a per-connection Socket.
// It runs the accept loop in a goroutine and returns immediately.
func (s *Socket) Serve(handler func(*Socket)) error {
    if s == nil || s.closed { return ErrClosed }
    if s.ln == nil || s.proto != TCP { return errors.New("Serve requires TCP listening socket") }
    go func() {
        for {
            c, err := s.ln.Accept()
            if err != nil { return }
            go func(conn net.Conn){ handler(&Socket{proto: TCP, conn: conn}) }(c)
        }
    }()
    return nil
}

// Read reads from the socket into p. For TCP it reads from the connected stream.
// For UDP it reads a single datagram into p (truncated if larger).
func (s *Socket) Read(p []byte) (int, error) {
    if s == nil || s.closed { return 0, ErrClosed }
    if s.conn != nil { return s.conn.Read(p) }
    if s.pc != nil {
        n, _, err := s.pc.ReadFrom(p)
        return n, err
    }
    return 0, errors.New("socket not readable")
}

// LocalAddr returns the local address string of the socket.
func (s *Socket) LocalAddr() string {
    if s == nil { return "" }
    if s.conn != nil && s.conn.LocalAddr() != nil { return s.conn.LocalAddr().String() }
    if s.pc != nil && s.pc.LocalAddr() != nil { return s.pc.LocalAddr().String() }
    if s.ln != nil && s.ln.Addr() != nil { return s.ln.Addr().String() }
    return ""
}

// RemoteAddr returns the remote address string for connected sockets (TCP). Empty for UDP listeners.
func (s *Socket) RemoteAddr() string {
    if s == nil { return "" }
    if s.conn != nil && s.conn.RemoteAddr() != nil { return s.conn.RemoteAddr().String() }
    return ""
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
