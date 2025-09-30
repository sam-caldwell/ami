package scheduler

import (
    "context"
    "sync"
    "time"
    "errors"
    "runtime"
)

// Policy enumerates worker scheduling strategies.
type Policy string

const (
    FIFO Policy = "fifo"
    LIFO Policy = "lifo"
    FAIR Policy = "fair"
    WORKSTEAL Policy = "worksteal"
)

// Config controls worker model per node kind.
type Config struct {
    Workers int
    Policy  Policy
    QueueCapacity int // 0=unbounded in practice
}

// Task represents a unit of work.
type Task struct {
    Source string
    Do func(context.Context)
}

// Pool executes tasks using the configured policy and worker limit.
type Pool struct {
    cfg Config
    ctx context.Context
    cancel context.CancelFunc

    wg sync.WaitGroup

    // queues for policies
    fifoCh chan Task
    lifoMu sync.Mutex
    lifo   []Task

    // fair policy: per-source queues
    fairMu sync.Mutex
    fairQ map[string][]Task
    fairOrder []string
    fairIdx int

    // worksteal: per-worker deques
    wsMu sync.Mutex
    wsQ  []chan Task
    wsRR int
}

var ErrInvalidWorkers = errors.New("workers must be >= 1")

func New(cfg Config) (*Pool, error) {
    if cfg.Workers <= 0 { return nil, ErrInvalidWorkers }
    if cfg.Policy == "" { cfg.Policy = FIFO }
    ctx, cancel := context.WithCancel(context.Background())
    p := &Pool{cfg: cfg, ctx: ctx, cancel: cancel}
    switch cfg.Policy {
    case FIFO:
        n := cfg.QueueCapacity; if n <= 0 { n = 0 }
        p.fifoCh = make(chan Task, n)
    case LIFO:
        p.lifo = make([]Task, 0)
    case FAIR:
        p.fairQ = make(map[string][]Task)
        p.fairOrder = make([]string, 0)
    case WORKSTEAL:
        p.wsQ = make([]chan Task, cfg.Workers)
        for i := 0; i < cfg.Workers; i++ {
            // small bounded channels to exercise stealing
            p.wsQ[i] = make(chan Task, 64)
        }
    }
    p.spawn()
    return p, nil
}

func (p *Pool) spawn() {
    for i := 0; i < p.cfg.Workers; i++ {
        p.wg.Add(1)
        wid := i
        go func(){ defer p.wg.Done(); p.runWorker(wid) }()
    }
}

func (p *Pool) Stop() { p.cancel(); p.wg.Wait() }

// Submit enqueues a task according to policy.
func (p *Pool) Submit(t Task) error {
    switch p.cfg.Policy {
    case FIFO:
        select { case p.fifoCh <- t: return nil; default: return errors.New("queue full") }
    case LIFO:
        p.lifoMu.Lock(); p.lifo = append(p.lifo, t); p.lifoMu.Unlock(); return nil
    case FAIR:
        p.fairMu.Lock()
        if _, ok := p.fairQ[t.Source]; !ok { p.fairOrder = append(p.fairOrder, t.Source) }
        p.fairQ[t.Source] = append(p.fairQ[t.Source], t)
        p.fairMu.Unlock()
        return nil
    case WORKSTEAL:
        // Try to enqueue into any worker's deque; spin briefly until accepted to avoid drops.
        p.wsMu.Lock()
        start := p.wsRR
        p.wsRR++
        n := len(p.wsQ)
        p.wsMu.Unlock()
        for {
            for i := 0; i < n; i++ {
                idx := (start + i) % n
                select { case p.wsQ[idx] <- t: return nil; default: }
            }
            runtime.Gosched()
            time.Sleep(1 * time.Millisecond)
        }
    default:
        return errors.New("unknown policy")
    }
}

func (p *Pool) runWorker(id int) {
    defer func(){ _ = recover() }()
    for {
        select {
        case <-p.ctx.Done(): return
        default:
        }
        var t Task
        ok := false
        switch p.cfg.Policy {
        case FIFO:
            select {
            case <-p.ctx.Done(): return
            case t, ok = <-p.fifoCh:
            }
        case LIFO:
            p.lifoMu.Lock()
            n := len(p.lifo)
            if n > 0 { t = p.lifo[n-1]; p.lifo = p.lifo[:n-1]; ok = true }
            p.lifoMu.Unlock()
            if !ok { runtime.Gosched(); time.Sleep(1 * time.Millisecond) }
        case FAIR:
            p.fairMu.Lock()
            if len(p.fairOrder) > 0 {
                src := p.fairOrder[p.fairIdx%len(p.fairOrder)]
                q := p.fairQ[src]
                if len(q) > 0 {
                    t = q[0]
                    p.fairQ[src] = q[1:]
                    ok = true
                }
                p.fairIdx++
            }
            p.fairMu.Unlock()
            if !ok { runtime.Gosched(); time.Sleep(1 * time.Millisecond) }
        case WORKSTEAL:
            // own queue fast path
            if t, ok = p.tryPopWS(id); !ok {
                // steal
                for i := 0; i < len(p.wsQ); i++ {
                    j := (id + i + 1) % len(p.wsQ)
                    if tt, ok2 := p.tryStealWS(j); ok2 { t = tt; ok = true; break }
                }
                if !ok { runtime.Gosched(); time.Sleep(1 * time.Millisecond) }
            }
        }
        if ok && t.Do != nil {
            t.Do(p.ctx)
        }
    }
}

func (p *Pool) tryPopWS(id int) (Task, bool) {
    select { case t := <-p.wsQ[id]: return t, true; default: return Task{}, false }
}
func (p *Pool) tryStealWS(id int) (Task, bool) { return p.tryPopWS(id) }
