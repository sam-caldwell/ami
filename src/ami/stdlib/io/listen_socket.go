package io

import (
    "errors"
    "net"
    "strconv"
)

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

