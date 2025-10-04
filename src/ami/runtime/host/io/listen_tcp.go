package io

// ListenTCP is a convenience wrapper for ListenSocket(TCP,...).
func ListenTCP(addr string, port uint16) (*Socket, error) { return ListenSocket(TCP, addr, port) }

