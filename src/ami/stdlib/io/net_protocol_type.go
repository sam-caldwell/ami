package io

// NetProtocol represents supported network protocols for sockets.
type NetProtocol string

const (
    TCP  NetProtocol = "TCP"
    UDP  NetProtocol = "UDP"
    ICMP NetProtocol = "ICMP"
)

