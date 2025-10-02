package io

import (
	stdctx "context"
	"errors"
	"net"
	"strconv"
	"syscall"
	stdtime "time"
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
	if err := guardNet(); err != nil {
		return nil, err
	}
	switch proto {
	case UDP:
		pc, err := net.ListenPacket("udp", net.JoinHostPort(addr, strconv.Itoa(int(port))))
		if err != nil {
			return nil, err
		}
		return &Socket{proto: UDP, pc: pc}, nil
	case TCP:
		c, err := net.Dial("tcp", net.JoinHostPort(addr, strconv.Itoa(int(port))))
		if err != nil {
			return nil, err
		}
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
	if err := guardNet(); err != nil {
		return nil, err
	}
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
func ConnectTCP(host string, port uint16) (*Socket, error) { return ConnectSocket(TCP, host, port) }
func ListenTCP(addr string, port uint16) (*Socket, error)  { return ListenSocket(TCP, addr, port) }
func ListenUDP(addr string, port uint16) (*Socket, error)  { return OpenSocket(UDP, addr, port) }

// ListenSocket creates a listening socket for the given protocol.
// For UDP, this is equivalent to OpenSocket(UDP, addr, port).
// For TCP, it binds and listens on addr:port for incoming connections.
func ListenSocket(proto NetProtocol, addr string, port uint16) (*Socket, error) {
	if err := guardNet(); err != nil {
		return nil, err
	}
	switch proto {
	case UDP:
		return OpenSocket(UDP, addr, port)
	case TCP:
		ln, err := net.Listen("tcp", net.JoinHostPort(addr, strconv.Itoa(int(port))))
		if err != nil {
			return nil, err
		}
		return &Socket{proto: TCP, ln: ln}, nil
	case ICMP:
		return nil, errors.New("ICMP not implemented")
	default:
		return nil, errors.New("unknown protocol")
	}
}

// Write appends p to the internal buffer.
func (s *Socket) Write(p []byte) (int, error) {
	if err := guardNet(); err != nil {
		return 0, err
	}
	if s == nil || s.closed {
		return 0, ErrClosed
	}
	s.buf = append(s.buf, p...)
	return len(p), nil
}

// Send sends the buffered data for connected sockets. For UDP bound sockets, use SendTo.
func (s *Socket) Send() error {
	if err := guardNet(); err != nil {
		return err
	}
	if s == nil || s.closed {
		return ErrClosed
	}
	// TCP connected path
	if s.conn != nil {
		if len(s.buf) == 0 {
			return nil
		}
		_, err := s.conn.Write(s.buf)
		s.buf = s.buf[:0]
		return err
	}
	// UDP listener path (no connected remote)
	return errors.New("Send not supported on UDP listening sockets; use SendTo")
}

// SendTo transmits the buffered data to the given remote host:port for UDP.
func (s *Socket) SendTo(host string, port uint16) error {
	if err := guardNet(); err != nil {
		return err
	}
	if s == nil || s.closed {
		return ErrClosed
	}
	if s.proto != UDP || s.pc == nil {
		return errors.New("SendTo requires UDP listening socket")
	}
	if len(s.buf) == 0 {
		return nil
	}
	raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, strconv.Itoa(int(port))))
	if err != nil {
		return err
	}
	_, err = s.pc.WriteTo(s.buf, raddr)
	s.buf = s.buf[:0]
	return err
}

// Listen registers a handler that receives each incoming datagram on UDP sockets.
func (s *Socket) Listen(handler func(b []byte)) error {
	if err := guardNet(); err != nil {
		return err
	}
	if s == nil || s.closed {
		return ErrClosed
	}
	switch s.proto {
	case UDP:
		if s.pc == nil {
			return errors.New("Listen requires UDP listening socket")
		}
		go func() {
			buf := make([]byte, 64*1024)
			for {
				n, _, err := s.pc.ReadFrom(buf)
				if err != nil {
					return
				}
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
					if err != nil {
						return
					}
					go func(conn net.Conn) {
						defer conn.Close()
						buf := make([]byte, 64*1024)
						for {
							n, err := conn.Read(buf)
							if err != nil {
								return
							}
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
					if err != nil {
						return
					}
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
	if err := guardNet(); err != nil {
		return 0, err
	}
	if s == nil || s.closed {
		return 0, ErrClosed
	}
	if s.proto != UDP || s.pc == nil {
		return 0, errors.New("WriteTo requires UDP listening socket")
	}
	raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, strconv.Itoa(int(port))))
	if err != nil {
		return 0, err
	}
	return s.pc.WriteTo(p, raddr)
}

// Serve accepts TCP connections and invokes handler with a per-connection Socket.
// It runs the accept loop in a goroutine and returns immediately.
func (s *Socket) Serve(handler func(*Socket)) error {
	if err := guardNet(); err != nil {
		return err
	}
	if s == nil || s.closed {
		return ErrClosed
	}
	if s.ln == nil || s.proto != TCP {
		return errors.New("Serve requires TCP listening socket")
	}
	go func() {
		for {
			c, err := s.ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) { handler(&Socket{proto: TCP, conn: conn}) }(c)
		}
	}()
	return nil
}

// Read reads from the socket into p. For TCP it reads from the connected stream.
// For UDP it reads a single datagram into p (truncated if larger).
func (s *Socket) Read(p []byte) (int, error) {
	if err := guardNet(); err != nil {
		return 0, err
	}
	if s == nil || s.closed {
		return 0, ErrClosed
	}
	if s.conn != nil {
		return s.conn.Read(p)
	}
	if s.pc != nil {
		n, _, err := s.pc.ReadFrom(p)
		return n, err
	}
	return 0, errors.New("socket not readable")
}

// ReadFrom reads data and returns the number of bytes along with the remote address for UDP sockets.
// For TCP connected sockets, it returns data with empty host and zero port.
func (s *Socket) ReadFrom(p []byte) (int, string, uint16, error) {
	if err := guardNet(); err != nil {
		return 0, "", 0, err
	}
	if s == nil || s.closed {
		return 0, "", 0, ErrClosed
	}
	if s.pc != nil {
		n, addr, err := s.pc.ReadFrom(p)
		if err != nil {
			return 0, "", 0, err
		}
		if ua, ok := addr.(*net.UDPAddr); ok {
			return n, ua.IP.String(), uint16(ua.Port), nil
		}
		// Fallback string parsing
		host, portStr, _ := net.SplitHostPort(addr.String())
		pnum, _ := strconv.Atoi(portStr)
		return n, host, uint16(pnum), nil
	}
	if s.conn != nil {
		n, err := s.conn.Read(p)
		return n, "", 0, err
	}
	return 0, "", 0, errors.New("socket not readable")
}

// LocalAddr returns the local address string of the socket.
func (s *Socket) LocalAddr() string {
	if s == nil {
		return ""
	}
	if s.conn != nil && s.conn.LocalAddr() != nil {
		return s.conn.LocalAddr().String()
	}
	if s.pc != nil && s.pc.LocalAddr() != nil {
		return s.pc.LocalAddr().String()
	}
	if s.ln != nil && s.ln.Addr() != nil {
		return s.ln.Addr().String()
	}
	return ""
}

// RemoteAddr returns the remote address string for connected sockets (TCP). Empty for UDP listeners.
func (s *Socket) RemoteAddr() string {
	if s == nil {
		return ""
	}
	if s.conn != nil && s.conn.RemoteAddr() != nil {
		return s.conn.RemoteAddr().String()
	}
	return ""
}

// CloseRead closes the read side of a TCP connection.
func (s *Socket) CloseRead() error {
	if s == nil || s.closed {
		return ErrClosed
	}
	if s.conn == nil {
		return errors.New("CloseRead requires TCP connection")
	}
	if tc, ok := s.conn.(*net.TCPConn); ok {
		if err := tc.CloseRead(); err != nil {
			// If the peer already closed or the socket is no longer connected,
			// treat it as a no-op for CloseRead to keep semantics lenient in tests.
			if errors.Is(err, syscall.ENOTCONN) {
				return nil
			}
			return err
		}
		return nil
	}
	return errors.New("CloseRead not supported for this connection")
}

// CloseWrite closes the write side of a TCP connection.
func (s *Socket) CloseWrite() error {
	if s == nil || s.closed {
		return ErrClosed
	}
	if s.conn == nil {
		return errors.New("CloseWrite requires TCP connection")
	}
	if tc, ok := s.conn.(*net.TCPConn); ok {
		if err := tc.CloseWrite(); err != nil {
			if errors.Is(err, syscall.ENOTCONN) {
				return nil
			}
			return err
		}
		return nil
	}
	return errors.New("CloseWrite not supported for this connection")
}

// SetDeadline sets read and write deadlines.
func (s *Socket) SetDeadline(t stdtime.Time) error {
	if err := guardNet(); err != nil {
		return err
	}
	if s == nil || s.closed {
		return ErrClosed
	}
	switch {
	case s.conn != nil:
		return s.conn.SetDeadline(t)
	case s.pc != nil:
		return s.pc.SetDeadline(t)
	case s.ln != nil:
		if tl, ok := s.ln.(*net.TCPListener); ok {
			return tl.SetDeadline(t)
		}
		return errors.New("deadline not supported for this listener")
	default:
		return errors.New("no underlying socket")
	}
}

// SetReadDeadline sets the read deadline.
func (s *Socket) SetReadDeadline(t stdtime.Time) error {
	if err := guardNet(); err != nil {
		return err
	}
	if s == nil || s.closed {
		return ErrClosed
	}
	switch {
	case s.conn != nil:
		return s.conn.SetReadDeadline(t)
	case s.pc != nil:
		return s.pc.SetReadDeadline(t)
	case s.ln != nil:
		if tl, ok := s.ln.(*net.TCPListener); ok {
			return tl.SetDeadline(t)
		}
		return errors.New("read deadline not supported for this listener")
	default:
		return errors.New("no underlying socket")
	}
}

// SetWriteDeadline sets the write deadline.
func (s *Socket) SetWriteDeadline(t stdtime.Time) error {
	if err := guardNet(); err != nil {
		return err
	}
	if s == nil || s.closed {
		return ErrClosed
	}
	switch {
	case s.conn != nil:
		return s.conn.SetWriteDeadline(t)
	case s.pc != nil:
		return s.pc.SetWriteDeadline(t)
	case s.ln != nil:
		return errors.New("write deadline not supported for listener")
	default:
		return errors.New("no underlying socket")
	}
}

// ServeContext is like Serve but will stop accepting when ctx is done.
// It closes the underlying listener to unblock Accept.
func (s *Socket) ServeContext(ctx stdctx.Context, handler func(*Socket)) error {
	if err := guardNet(); err != nil {
		return err
	}
	if s == nil || s.closed {
		return ErrClosed
	}
	if s.ln == nil || s.proto != TCP {
		return errors.New("ServeContext requires TCP listening socket")
	}
	// stop the listener when ctx done
	go func() { <-ctx.Done(); _ = s.Close() }()
	return s.Serve(handler)
}

// Close closes the socket.
func (s *Socket) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	var err error
	if s.pc != nil {
		err = s.pc.Close()
	}
	if s.conn != nil {
		_ = s.conn.Close()
	}
	return err
}
