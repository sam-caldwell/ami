package tester

// Fixture describes a permitted file path and access mode for a test case.
type Fixture struct {
    Path string
    Mode string // ro|rw
}

