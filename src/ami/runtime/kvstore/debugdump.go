package kvstore

import (
    "encoding/json"
    "time"
)

// DebugDump returns a human-readable summary of the store's contents.
func (s *Store) DebugDump() string {
    s.mu.Lock()
    defer s.mu.Unlock()
    type dumpEntry struct {
        Key            string `json:"key"`
        ExpiresAt      string `json:"expiresAt,omitempty"`
        Sliding        bool   `json:"sliding,omitempty"`
        ReadsRemaining int    `json:"readsRemaining,omitempty"`
        Size           int64  `json:"size"`
        LastAccess     string `json:"lastAccess"`
    }
    tmp := struct {
        Entries   int         `json:"entries"`
        BytesUsed int64       `json:"bytesUsed"`
        Items     []dumpEntry `json:"items"`
    }{Entries: len(s.items), BytesUsed: s.used}
    tmp.Items = make([]dumpEntry, 0, len(s.items))
    for _, e := range s.items {
        de := dumpEntry{Key: e.key, Size: e.size, ReadsRemaining: e.readsRemaining, LastAccess: e.lastAccess.Format(time.RFC3339Nano)}
        if !e.expiresAt.IsZero() {
            de.ExpiresAt = e.expiresAt.Format(time.RFC3339Nano)
        }
        if e.sliding {
            de.Sliding = true
        }
        tmp.Items = append(tmp.Items, de)
    }
    b, _ := json.Marshal(tmp)
    return string(b)
}

