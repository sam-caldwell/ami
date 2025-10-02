package merge

import "time"

type partition struct{
    buf []item
    seen map[string]struct{}
    last time.Time
}

