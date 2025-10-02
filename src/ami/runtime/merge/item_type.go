package merge

import ev "github.com/sam-caldwell/ami/src/schemas/events"

type item struct{
    ev ev.Event
    keys []any // extracted sort key values
    seq int64
    key any
}

