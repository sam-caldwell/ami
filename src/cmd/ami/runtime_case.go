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
}

type runtimeCase struct{
    File string // relative path
    Name string
    Spec runtimeSpec
}

