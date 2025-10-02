package io

// NetInterface is a minimal description of a network interface.
type NetInterface struct {
    Index int
    Name  string
    MTU   int
    Flags string
    Addrs []string
}

