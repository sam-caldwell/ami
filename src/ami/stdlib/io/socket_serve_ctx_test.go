package io

import (
    stdctx "context"
    "net"
    "strconv"
    "testing"
    "time"
)

func TestServeContext_CancelBeforeConnect_NoPanic(t *testing.T) {
    srv, err := ListenTCP("127.0.0.1", 0)
    if err != nil { t.Fatalf("listen tcp: %v", err) }
    defer srv.Close()

    ctx, cancel := stdctx.WithCancel(stdctx.Background())
    cancel() // cancel immediately
    if err := srv.ServeContext(ctx, func(c *Socket){ _ = c.Close() }); err != nil { t.Fatalf("ServeContext: %v", err) }

    host, portStr, _ := net.SplitHostPort(srv.LocalAddr())
    p, _ := strconv.Atoi(portStr)
    // Attempt a connection; allow either connect failure, or success but no handler invocation
    c, _ := ConnectTCP(host, uint16(p))
    if c != nil { _ = c.Close() }
    // Allow a moment; core assertion is that ServeContext handled cancel without panic or deadlock
    time.Sleep(20 * time.Millisecond)
}
