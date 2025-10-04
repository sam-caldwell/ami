package exec

import (
    "time"
    amiio "github.com/sam-caldwell/ami/src/ami/runtime/host/io"
    events "github.com/sam-caldwell/ami/src/schemas/events"
    errs "github.com/sam-caldwell/ami/src/schemas/errors"
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
    // Worker registry for Transform stages. Keyed by worker name (e.g., "W").
    // The function may return either an Event (already-wrapped) or a bare payload (docx-aligned), with error.
    Workers       map[string]func(e events.Event) (any, error)
    // ErrorChan: when provided, worker errors are sent here as errors.v1
    // payloads instead of being injected into the main event stream.
    // Callers should drain this channel to avoid goroutine leaks.
    ErrorChan     chan errs.Error
}
