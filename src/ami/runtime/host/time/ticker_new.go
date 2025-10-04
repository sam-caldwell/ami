package amitime

// NewTicker constructs a Ticker with period d.
func NewTicker(d Duration) *Ticker { return &Ticker{d: d} }

