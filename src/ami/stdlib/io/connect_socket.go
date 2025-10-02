package io

import "errors"

// ConnectSocket establishes a connected socket when supported.
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

