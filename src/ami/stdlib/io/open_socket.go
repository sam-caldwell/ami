package io

import (
    "errors"
    "net"
    "strconv"
)

// OpenSocket opens a UDP socket bound to addr:port, or connects TCP.
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

