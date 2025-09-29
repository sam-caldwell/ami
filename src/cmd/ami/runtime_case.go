package main

type fixtureSpec struct{
    Path string
    Mode string // ro|rw
}

type runtimeSpec struct{
    InputJSON   string
    ExpectJSON  string
    ExpectError string // expected error code (optional)
    TimeoutMs   int    // per-case override
    Fixtures    []fixtureSpec
    SkipReason  string
    // KV harness integration
    KvNS    string            // namespace "pipeline/node" or arbitrary
    KvPut   map[string]string // key=value pairs
    KvGet   []string          // keys to fetch (for side effects)
    KvEmit  bool              // emit per-case metrics/dump
}

type runtimeCase struct{
    File string // relative path
    Name string
    Spec runtimeSpec
}
