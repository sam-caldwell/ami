package merge

import (
    "errors"
    "time"
    "sort"
    "strings"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
    "strconv"
)

// Operator merges events according to Plan. It is synchronous and thread-safe
// for Push/Pop when called by a single producer/consumer pair. For multi-
// producer, wrap Push with external synchronization.
type Operator struct{
    plan Plan
    parts map[string]*partition
    rr []string
    rrIdx int
    seq int64
}

type item struct{
    ev ev.Event
    keys []any // extracted sort key values
    seq int64
}

type partition struct{
    buf []item
    seen map[string]struct{}
    last time.Time
}

var ErrBackpressure = errors.New("merge buffer full")

func NewOperator(p Plan) *Operator {
    return &Operator{plan:p, parts: map[string]*partition{}, rr: make([]string, 0)}
}

func (op *Operator) partitionKey(e ev.Event) string {
    if op.plan.PartitionBy == "" { return "__default__" }
    if v, ok := extractPath(e.Payload, op.plan.PartitionBy); ok { return toKey(v) }
    return "__default__"
}

func (op *Operator) dedupKey(e ev.Event) (string, bool) {
    field := op.plan.Dedup.Field
    if field == "" { field = op.plan.Key }
    if field == "" { return "", false }
    if v, ok := extractPath(e.Payload, field); ok { return toKey(v), true }
    return "", false
}

func (op *Operator) Push(e ev.Event) error {
    pk := op.partitionKey(e)
    part := op.parts[pk]
    if part == nil { part = &partition{buf: make([]item,0), seen: map[string]struct{}{}, last: time.Now()}; op.parts[pk]=part; op.rr = append(op.rr, pk) }
    // watermark filter
    if op.plan.Watermark != nil && op.plan.Watermark.Field != "" {
        if v, ok := extractPath(e.Payload, op.plan.Watermark.Field); ok {
            if t, ok2 := toTime(v); ok2 {
                // Any event older than now - lateness is dropped
                if op.plan.Watermark.LatenessMs > 0 {
                    if t.Before(time.Now().Add(-time.Duration(op.plan.Watermark.LatenessMs) * time.Millisecond)) {
                        return nil
                    }
                }
            }
        }
    }
    // dedup
    if dk, ok := op.dedupKey(e); ok {
        if _, seen := part.seen[dk]; seen { return nil }
        part.seen[dk] = struct{}{}
    }
    // backpressure/window
    cap := op.plan.Buffer.Capacity
    if op.plan.Window > 0 && (cap == 0 || op.plan.Window < cap) { cap = op.plan.Window }
    if cap > 0 && len(part.buf) >= cap {
        switch op.plan.Buffer.Policy {
        case "block":
            return ErrBackpressure
        case "dropOldest", "shuntOldest":
            if len(part.buf) > 0 {
                part.buf = part.buf[1:]
            }
        case "dropNewest", "shuntNewest":
            return nil
        default:
            return nil
        }
    }
    // extract sort keys
    keys := make([]any, len(op.plan.Sort))
    for i, k := range op.plan.Sort {
        if v, ok := extractPath(e.Payload, k.Field); ok { keys[i] = v } else { keys[i] = nil }
    }
    op.seq++
    it := item{ev: e, keys: keys, seq: op.seq}
    part.buf = append(part.buf, it)
    // maintain ordering on insert
    sort.SliceStable(part.buf, func(i, j int) bool { return less(part.buf[i], part.buf[j], op.plan) })
    part.last = time.Now()
    return nil
}

// Pop returns the next event from a non-empty partition, using round-robin for fairness across partitions.
func (op *Operator) Pop() (ev.Event, bool) {
    if len(op.rr) == 0 { return ev.Event{}, false }
    for k := 0; k < len(op.rr); k++ {
        idx := (op.rrIdx + k) % len(op.rr)
        pk := op.rr[idx]
        part := op.parts[pk]
        if part == nil || len(part.buf) == 0 { continue }
        it := part.buf[0]
        part.buf = part.buf[1:]
        op.rrIdx = (idx + 1) % len(op.rr)
        return it.ev, true
    }
    return ev.Event{}, false
}

// ExpireStale applies TimeoutMs by clearing buffers for partitions that have
// been idle since before now-Timeout. Returns the number of dropped items.
func (op *Operator) ExpireStale(now time.Time) int {
    if op.plan.TimeoutMs <= 0 { return 0 }
    var dropped int
    cutoff := now.Add(-time.Duration(op.plan.TimeoutMs) * time.Millisecond)
    for _, pk := range op.rr {
        part := op.parts[pk]
        if part == nil { continue }
        if part.last.Before(cutoff) && len(part.buf) > 0 {
            dropped += len(part.buf)
            part.buf = part.buf[:0]
            // reset seen map only if dedup is enabled
            if op.plan.Dedup.Field != "" || op.plan.Key != "" { part.seen = map[string]struct{}{} }
            part.last = now
        }
    }
    return dropped
}

func less(a, b item, p Plan) bool {
    for i, k := range p.Sort {
        av := a.keys[i]
        bv := b.keys[i]
        c := cmp(av, bv)
        if c == 0 { continue }
        if k.Order == "desc" { return c > 0 }
        return c < 0
    }
    if p.Stable { return a.seq < b.seq }
    return false
}

func cmp(a, b any) int {
    switch av := a.(type) {
    case int:
        bv, ok := b.(int); if !ok { return 0 }
        switch { case av < bv: return -1; case av > bv: return 1; default: return 0 }
    case int64:
        switch bv := b.(type) {
        case int64:
            if av < bv {return -1} else if av > bv {return 1}; return 0
        case int:
            ai := av; bi := int64(bv)
            if ai < bi {return -1} else if ai > bi {return 1}; return 0
        default:
            return 0
        }
    case float64:
        bv, ok := b.(float64); if !ok { return 0 }
        switch { case av < bv: return -1; case av > bv: return 1; default: return 0 }
    case string:
        bv, ok := b.(string); if !ok { return 0 }
        switch { case av < bv: return -1; case av > bv: return 1; default: return 0 }
    case time.Time:
        bv, ok := b.(time.Time); if !ok { return 0 }
        switch { case av.Before(bv): return -1; case av.After(bv): return 1; default: return 0 }
    default:
        return 0
    }
}

// extractPath reads dotted path from JSON-like payloads represented as map[string]any.
func extractPath(root any, path string) (any, bool) {
    if path == "" { return root, true }
    m, ok := root.(map[string]any)
    if !ok { return nil, false }
    cur := any(m)
    for _, seg := range strings.Split(path, ".") {
        mm, ok := cur.(map[string]any)
        if !ok { return nil, false }
        v, ok := mm[seg]
        if !ok { return nil, false }
        cur = v
    }
    return cur, true
}

func toKey(v any) string {
    switch x := v.(type) {
    case string: return x
    case int: return itoa(int64(x))
    case int64: return itoa(x)
    case float64: return ftoa(x)
    case bool: if x { return "true" } else { return "false" }
    default: return "" // non-deterministic; avoid map/array stringification here
    }
}

func toTime(v any) (time.Time, bool) {
    switch x := v.(type) {
    case time.Time: return x, true
    case string:
        // try RFC3339 subset
        if t, err := time.Parse(time.RFC3339, x); err == nil { return t, true }
        // try unix seconds
        return time.Unix(0,0), false
    default:
        return time.Unix(0,0), false
    }
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
func ftoa(f float64) string { return strconv.FormatFloat(f, 'g', -1, 64) }
