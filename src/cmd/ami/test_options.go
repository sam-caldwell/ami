package main

// TestOptions holds runtime-testing related options configured by CLI flags.
type TestOptions struct {
    TimeoutMs   int
    Parallel    int
    Failfast    bool
    RunPattern  string
}

var currentTestOptions TestOptions

func setTestOptions(o TestOptions) { currentTestOptions = o }

