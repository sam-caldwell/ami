package io

import "net"

// Interfaces lists available network interfaces.
func Interfaces() ([]NetInterface, error) {
    if err := guardNet(); err != nil { return nil, err }
    ifs, err := net.Interfaces()
    if err != nil { return nil, err }
    out := make([]NetInterface, 0, len(ifs))
    for _, iface := range ifs {
        addrs, _ := iface.Addrs()
        as := make([]string, 0, len(addrs))
        for _, a := range addrs { as = append(as, a.String()) }
        out = append(out, NetInterface{Index: iface.Index, Name: iface.Name, MTU: iface.MTU, Flags: iface.Flags.String(), Addrs: as})
    }
    return out, nil
}

