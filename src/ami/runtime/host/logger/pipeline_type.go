package logger

import (
    "errors"
    "sync"
    "time"
)

// Pipeline provides buffered, batched writes to a sink with backpressure policy.
type Pipeline struct {
    cfg    Config
    ch     chan []byte
    stop   chan struct{}
    wg     sync.WaitGroup
    mu     sync.Mutex
    buffer [][]byte
    // counters (protected by mu for determinism in tests)
    enqueued int64
    written  int64
    dropped  int64
    batches  int64
    flushes  int64
}

// Start begins consuming from the queue and writing to the sink.
func (p *Pipeline) Start() error {
    if p.cfg.Sink == nil { return errors.New("nil sink") }
    if err := p.cfg.Sink.Start(); err != nil { return err }
    p.wg.Add(1)
    go p.run()
    return nil
}

func (p *Pipeline) run() {
    defer p.wg.Done()
    var ticker *time.Ticker
    if p.cfg.FlushInterval > 0 {
        ticker = time.NewTicker(p.cfg.FlushInterval)
        defer ticker.Stop()
    }
    for {
        select {
        case <-p.stop:
            // Drain any pending items from the channel before final flush.
            for {
                select {
                case b := <-p.ch:
                    p.append(b)
                default:
                    goto drained
                }
            }
        drained:
            p.flush()
            _ = p.cfg.Sink.Close()
            return
        case b := <-p.ch:
            p.append(b)
            if p.cfg.BatchMax > 0 && len(p.buffer) >= p.cfg.BatchMax {
                p.flush()
            }
        default:
            if ticker == nil {
                // Avoid busy loop when no ticker; block until next event
                select {
                case <-p.stop:
                    for {
                        select {
                        case b := <-p.ch: p.append(b)
                        default: goto drained2
                        }
                    }
                drained2:
                    p.flush(); _ = p.cfg.Sink.Close(); return
                case b := <-p.ch:
                    p.append(b)
                    if p.cfg.BatchMax > 0 && len(p.buffer) >= p.cfg.BatchMax { p.flush() }
                }
            } else {
                select {
                case <-p.stop:
                    for {
                        select {
                        case b := <-p.ch: p.append(b)
                        default: goto drained3
                        }
                    }
                drained3:
                    p.flush(); _ = p.cfg.Sink.Close(); return
                case <-ticker.C:
                    p.flush()
                case b := <-p.ch:
                    p.append(b)
                    if p.cfg.BatchMax > 0 && len(p.buffer) >= p.cfg.BatchMax { p.flush() }
                }
            }
        }
    }
}

func (p *Pipeline) append(b []byte) {
    p.mu.Lock()
    p.buffer = append(p.buffer, b)
    p.mu.Unlock()
}

func (p *Pipeline) flush() {
    p.mu.Lock()
    if len(p.buffer) == 0 { p.mu.Unlock(); return }
    batch := p.buffer
    p.buffer = nil
    p.batches++
    p.flushes++
    p.mu.Unlock()
    for _, line := range batch {
        if len(p.cfg.JSONRedactKeys) > 0 || len(p.cfg.JSONRedactPrefixes) > 0 || len(p.cfg.JSONAllowKeys) > 0 || len(p.cfg.JSONDenyKeys) > 0 {
            if r, ok := redactLogV1LineAdvanced(line, p.cfg.JSONAllowKeys, p.cfg.JSONDenyKeys, p.cfg.JSONRedactKeys, p.cfg.JSONRedactPrefixes); ok {
                line = r
            }
        }
        _ = p.cfg.Sink.Write(line)
        p.mu.Lock(); p.written++; p.mu.Unlock()
    }
}

// Enqueue writes a line into the pipeline according to backpressure policy.
func (p *Pipeline) Enqueue(line []byte) error {
    switch p.cfg.Policy {
    case Block:
        p.ch <- line
        p.mu.Lock(); p.enqueued++; p.mu.Unlock()
        return nil
    case DropNewest:
        select {
        case p.ch <- line:
            p.mu.Lock(); p.enqueued++; p.mu.Unlock()
            return nil
        default:
            p.mu.Lock(); p.dropped++; p.mu.Unlock(); return ErrDropped
        }
    case DropOldest:
        select {
        case p.ch <- line:
            p.mu.Lock(); p.enqueued++; p.mu.Unlock(); return nil
        default:
            // drop one oldest if possible, then try enqueue
            select {
            case <-p.ch:
                p.mu.Lock(); p.dropped++; p.mu.Unlock()
            default:
                // nothing to drop
            }
            select {
            case p.ch <- line:
                p.mu.Lock(); p.enqueued++; p.mu.Unlock(); return nil
            default:
                p.mu.Lock(); p.dropped++; p.mu.Unlock(); return ErrDropped
            }
        }
    default:
        // default to block
        p.ch <- line
        p.mu.Lock(); p.enqueued++; p.mu.Unlock()
        return nil
    }
}

// Close stops the pipeline and flushes remaining data.
func (p *Pipeline) Close() {
    close(p.stop)
    p.wg.Wait()
}

// Stats returns a snapshot of pipeline counters.
func (p *Pipeline) Stats() Stats {
    p.mu.Lock(); defer p.mu.Unlock()
    return Stats{Enqueued: p.enqueued, Written: p.written, Dropped: p.dropped, Batches: p.batches, Flushes: p.flushes}
}
