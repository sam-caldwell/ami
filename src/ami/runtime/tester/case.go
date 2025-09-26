package tester

// Case describes a Phase 2 runtime test case for a compiled AMI pipeline.
type Case struct {
    Name        string
    Pipeline    string // pipeline name or entry
    InputJSON   string // input payload serialized as JSON
    ExpectJSON  string // expected output payload as JSON
    ExpectError string // expected runtime error code (optional)
    TimeoutMs   int    // per-case timeout
    Fixtures    []Fixture
}

