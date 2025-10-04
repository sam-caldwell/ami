package amsignal

// SignalType is an enum of OS signals supported by Register.
type SignalType int

const (
    SIGINT SignalType = iota + 1
    SIGTERM
    SIGHUP
    SIGQUIT
)
