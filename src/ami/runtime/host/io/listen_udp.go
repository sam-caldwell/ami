package io

// ListenUDP is a convenience wrapper for OpenSocket(UDP,...).
func ListenUDP(addr string, port uint16) (*Socket, error) { return OpenSocket(UDP, addr, port) }

