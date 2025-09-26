package kvstore

// StoreInfo captures a snapshot of a registered store for observability.
type StoreInfo struct {
    Pipeline string
    Node     string
    Stats    Stats
    Dump     string
}

