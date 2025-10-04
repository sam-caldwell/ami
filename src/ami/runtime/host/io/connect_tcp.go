package io

// ConnectTCP is a convenience wrapper for ConnectSocket(TCP,...).
func ConnectTCP(host string, port uint16) (*Socket, error) { return ConnectSocket(TCP, host, port) }

