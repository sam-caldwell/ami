package exec

import (
    "time"
    amiio "github.com/sam-caldwell/ami/src/ami/runtime/host/io"
)

// ExecOptions control runner behavior for sources/sinks in simulation.
type ExecOptions struct {
    SourceType    string        // auto|file|timer
    TimerInterval time.Duration // used when SourceType=timer or auto+Timer node
    TimerCount    int           // number of timer events (0=unlimited)
    Sandbox       SandboxPolicy // source/ingress sandbox capability policy
    // Network source configuration (when SourceType=net.tcp, net.udp)
    NetProtocol   amiio.NetProtocol
    NetAddr       string
    NetPort       uint16
}
